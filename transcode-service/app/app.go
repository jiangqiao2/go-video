package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	transcodepb "go-vedio-1/proto/transcode"
	"google.golang.org/grpc"

	transcodeGrpc "transcode-service/ddd/adapter/grpc"
	app "transcode-service/ddd/application/app"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/registry"
	"transcode-service/pkg/repository"
	"transcode-service/pkg/utils"

	"github.com/gin-gonic/gin"

	_ "transcode-service/ddd/adapter/http"
	_ "transcode-service/ddd/infrastructure/worker"
	// 导入资源和模块包以触发init函数
	_ "transcode-service/internal/resource"
)

func Run() {
	// 先使用标准输出确保能看到日志
	fmt.Println("[STARTUP] 开始启动转码服务...")

	// 加载配置
	fmt.Println("[STARTUP] 正在加载配置文件...")
	cfg, err := config.Load("configs/config.dev.yaml")
	if err != nil {
		fmt.Printf("[ERROR] 加载配置失败: %v\n", err)
		os.Exit(1)
	}
	// 设置全局配置（必须在资源管理器初始化之前）
	config.SetGlobalConfig(cfg)
	fmt.Println("[STARTUP] 配置文件加载成功")

	// 立即初始化日志服务（确保所有后续组件都能使用正确的日志器）
	fmt.Println("[STARTUP] 正在初始化日志服务...")
	logService := logger.NewLogger(cfg)
	logger.SetGlobalLogger(logService)
	fmt.Println("[STARTUP] 日志服务初始化完成")

	// 验证日志器配置
	logger.Debug("日志器初始化完成", map[string]interface{}{
		"level":  cfg.Log.Level,
		"format": cfg.Log.Format,
		"output": cfg.Log.Output,
	})

	logger.Info("转码服务启动", map[string]interface{}{"version": "1.0.0", "env": "development"})

	// 资源管理器初始化
	logger.Info("正在初始化资源管理器...")
	manager.MustInitResources()
	defer manager.CloseResources()
	logger.Info("资源管理器初始化完成")

	// 初始化数据库（用于依赖注入）
	logger.Info("正在初始化数据库连接...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("初始化数据库失败", map[string]interface{}{"error": err})
	}
	defer db.Close()
	logger.Info("数据库连接成功")

	// 初始化JWT工具
	logger.Info("正在初始化JWT工具...")
	jwtUtil := utils.DefaultJWTUtil()
	logger.Info("JWT工具初始化成功")

	// 初始化转码相关组件
	logger.Info("正在初始化转码相关组件...")

	// 初始化应用服务
	transcodeAppService := app.DefaultTranscodeApp()
	logger.Info("转码相关组件初始化完成")

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:      db.Self,
		Config:  cfg,
		JWTUtil: jwtUtil,
	}

	// 初始化所有服务
	logger.Info("正在初始化所有服务...")
	manager.MustInitServices(deps)
	logger.Info("所有服务初始化完成")

	// 初始化所有组件
	logger.Info("正在初始化所有组件...")
	manager.MustInitComponents(deps)
	logger.Info("所有组件初始化完成")

	// 转码服务组件初始化完成
	logger.Info("转码服务组件初始化完成")

	var (
		serviceRegistry *registry.ServiceRegistry
		grpcListener    net.Listener
		grpcServer      *grpc.Server
		serviceConfig   registry.ServiceConfig
	)

	// 启动gRPC服务
	logger.Info("正在启动gRPC服务器...")
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPCServer.Host, cfg.GRPCServer.Port)
	grpcListener, err = net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("监听gRPC端口失败", map[string]interface{}{
			"address": grpcAddr,
			"error":   err,
		})
	}

	grpcServer = grpc.NewServer()
	transcodepb.RegisterTranscodeServiceServer(
		grpcServer,
		transcodeGrpc.NewTranscodeGrpcServer(transcodeAppService),
	)

	go func() {
		logger.Info("gRPC服务器已启动", map[string]interface{}{
			"address": grpcAddr,
			"service": "transcode-service",
		})
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("gRPC服务器运行异常", map[string]interface{}{
				"error": err,
			})
		}
	}()

	// 注册服务到etcd（可配置关闭）
	if cfg.ServiceRegistry.Enabled && len(cfg.Etcd.Endpoints) > 0 {
		logger.Info("正在注册服务到etcd...")
		registryConfig := registry.RegistryConfig{
			Endpoints:      cfg.Etcd.Endpoints,
			DialTimeout:    cfg.Etcd.DialTimeout,
			RequestTimeout: cfg.Etcd.RequestTimeout,
			Username:       cfg.Etcd.Username,
			Password:       cfg.Etcd.Password,
		}
		serviceConfig = registry.ServiceConfig{
			ServiceName:     cfg.ServiceRegistry.ServiceName,
			ServiceID:       cfg.ServiceRegistry.ServiceID,
			TTL:             cfg.ServiceRegistry.TTL,
			RefreshInterval: cfg.ServiceRegistry.RefreshInterval,
		}

		registerHost := cfg.ServiceRegistry.RegisterHost
		if registerHost == "" {
			registerHost = cfg.GRPCServer.Host
			if registerHost == "" || registerHost == "0.0.0.0" {
				registerHost = "localhost"
			}
		}
		serviceAddr := fmt.Sprintf("%s:%d", registerHost, cfg.GRPCServer.Port)

		serviceRegistry, err = registry.NewServiceRegistry(registryConfig, serviceConfig, serviceAddr)
		if err != nil {
			logger.Fatal("创建服务注册失败", map[string]interface{}{
				"error": err,
			})
		}
		if err := serviceRegistry.Register(); err != nil {
			logger.Fatal("服务注册到etcd失败", map[string]interface{}{
				"error": err,
			})
		}
		logger.Info("服务注册成功", map[string]interface{}{
			"service": serviceConfig.ServiceName,
			"address": serviceAddr,
		})
	} else {
		logger.Info("跳过etcd服务注册（未开启或未配置etcd）")
	}

	// 创建Gin引擎
	logger.Info("正在创建HTTP路由...")
	router := gin.Default()

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "transcode-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 注册所有路由
	logger.Info("正在注册所有路由...")
	manager.RegisterAllRoutes(router)
	logger.Info("路由注册完成")

	// 启动HTTP服务器
	port := getEnv("PORT", "8083")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("启动服务器失败", map[string]interface{}{"error": err})
		}
	}()

	logger.Info("HTTP服务器启动成功", map[string]interface{}{
		"port":       port,
		"service":    "transcode-service",
		"health_url": fmt.Sprintf("http://localhost:%s/health", port),
		"api_url":    fmt.Sprintf("http://localhost:%s/api/v1", port),
	})

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到关闭信号，正在优雅关闭服务器...")

	if serviceRegistry != nil {
		logger.Info("正在注销服务注册...", map[string]interface{}{
			"service": serviceConfig.ServiceName,
		})
		if err := serviceRegistry.Deregister(); err != nil {
			logger.Error("注销服务注册失败", map[string]interface{}{
				"error": err,
			})
		} else {
			logger.Info("服务注册已注销", map[string]interface{}{
				"service": serviceConfig.ServiceName,
			})
		}
	}

	if grpcServer != nil {
		logger.Info("正在停止gRPC服务器...", map[string]interface{}{
			"address": grpcAddr,
		})
		grpcServer.GracefulStop()
	}

	// 关闭所有组件
	logger.Info("正在关闭所有组件...")
	manager.Shutdown()
	logger.Info("所有组件已关闭")

	// 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", map[string]interface{}{"error": err})
	}

	logger.Info("服务器已安全退出")

	// 关闭日志服务
	logger.Info("正在关闭日志服务...")
	if logService != nil {
		logService.Close()
	}

	fmt.Println("[SHUTDOWN] 转码服务已安全退出")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
