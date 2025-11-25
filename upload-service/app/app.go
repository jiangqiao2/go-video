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
	"upload-service/ddd/adapter/task"

	"github.com/gin-gonic/gin"

	uploadpb "go-vedio-1/proto/upload"
	"google.golang.org/grpc"

	uploadGrpc "upload-service/ddd/adapter/grpc"
	service "upload-service/ddd/domain/service"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	"upload-service/pkg/config"
	"upload-service/pkg/logger"
	"upload-service/pkg/manager"
	"upload-service/pkg/registry"
	"upload-service/pkg/repository"
	"upload-service/pkg/utils"

	_ "upload-service/ddd/adapter/http"
	// 导入资源和模块包以触发init函数
	_ "upload-service/internal/resource"
)

func Run() {
	// 先使用标准输出确保能看到日志
	fmt.Println("[STARTUP] 开始启动上传服务...")

	// 加载配置
	fmt.Println("[STARTUP] 正在加载配置文件...")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.dev.yaml"
	}
	cfg, err := config.Load(configPath)
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

	logger.Info("上传服务启动", map[string]interface{}{"version": "1.0.0", "env": "development"})

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

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:      db.Self,
		Config:  cfg,
		JWTUtil: jwtUtil,
	}

	var (
		grpcListener     net.Listener
		grpcServer       *grpc.Server
		grpcRegistry     *registry.ServiceRegistry
		grpcAddr         string
		registryCfg      registry.RegistryConfig
		serviceDiscovery *registry.ServiceDiscovery
	)

	useServiceRegistry := cfg.ServiceRegistry.Enabled && len(cfg.Etcd.Endpoints) > 0
	useGrpcRegistry := cfg.GRPCServiceRegistry.Enabled && len(cfg.Etcd.Endpoints) > 0
	if useServiceRegistry || useGrpcRegistry {
		registryCfg = registry.RegistryConfig{
			Endpoints:      cfg.Etcd.Endpoints,
			DialTimeout:    cfg.Etcd.DialTimeout,
			RequestTimeout: cfg.Etcd.RequestTimeout,
			Username:       cfg.Etcd.Username,
			Password:       cfg.Etcd.Password,
		}
	}

	if useServiceRegistry {
		logger.Info("正在初始化服务发现...")
		serviceDiscovery, err = registry.NewServiceDiscovery(registryCfg)
		if err != nil {
			logger.Warn("创建服务发现失败，降级为直连模式", map[string]interface{}{"error": err.Error()})
			serviceDiscovery = nil
			useServiceRegistry = false
		} else {
			defer func() {
				if err := serviceDiscovery.Close(); err != nil {
					logger.Warn("关闭服务发现失败", map[string]interface{}{"error": err.Error()})
				}
			}()
			logger.Info("服务发现初始化完成")

			userSvcName := cfg.Dependencies.UserService.ServiceName
			if userSvcName == "" {
				userSvcName = "user-service"
			}
			serviceDiscovery.WatchService(userSvcName)
			logger.Info("开始监听user-service服务变化")

			transcodeSvcName := cfg.Dependencies.TranscodeService.ServiceName
			if transcodeSvcName == "" {
				transcodeSvcName = "transcode-service"
			}
			serviceDiscovery.WatchService(transcodeSvcName)
			logger.Info("开始监听transcode-service服务变化")
		}
	} else {
		logger.Info("跳过服务发现（未开启或未配置etcd），使用直连配置")
	}

	// 初始化gRPC客户端
	logger.Info("正在初始化gRPC客户端...")
	clientConfig := grpcClient.ClientConfig{
		Timeout:        cfg.GRPC.Timeout,
		MaxRecvMsgSize: cfg.GRPC.MaxRecvMsgSize,
		MaxSendMsgSize: cfg.GRPC.MaxSendMsgSize,
		RetryTimes:     cfg.GRPC.RetryTimes,
	}
	userServiceClient, err := grpcClient.NewUserServiceClient(serviceDiscovery, clientConfig)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create gRPC client: %v", err))
		return
	}
	logger.Info("gRPC客户端初始化完成")

	// 注册upload-service到etcd
	logger.Info("正在注册upload-service到etcd...")
	if useServiceRegistry {
		uploadServiceConfig := registry.ServiceConfig{
			ServiceName:     cfg.ServiceRegistry.ServiceName,
			ServiceID:       cfg.ServiceRegistry.ServiceID,
			TTL:             cfg.ServiceRegistry.TTL,
			RefreshInterval: cfg.ServiceRegistry.RefreshInterval,
		}
		registerHost := cfg.ServiceRegistry.RegisterHost
		if registerHost == "" {
			registerHost = cfg.Server.Host
			if registerHost == "" || registerHost == "0.0.0.0" {
				registerHost = "localhost"
			}
		}
		httpAddr := fmt.Sprintf("%s:%d", registerHost, cfg.Server.Port)
		uploadRegistry, err := registry.NewServiceRegistry(registryCfg, uploadServiceConfig, httpAddr)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to create upload service registry: %v", err))
			return
		}
		if err := uploadRegistry.Register(); err != nil {
			logger.Fatal(fmt.Sprintf("Failed to register upload service: %v", err))
			return
		}
		logger.Info("upload-service注册到etcd成功")
	} else {
		logger.Info("跳过upload-service注册（未开启或未配置etcd）")
	}

	// 将gRPC客户端添加到依赖中
	deps.UserServiceClient = userServiceClient

	// 初始化所有服务（在gRPC客户端初始化之后）
	logger.Info("正在初始化所有服务...")
	manager.MustInitServices(deps)
	logger.Info("所有服务初始化完成")

	// 初始化所有组件
	logger.Info("正在初始化所有组件...")
	manager.MustInitComponents(deps)
	logger.Info("所有组件初始化完成")

	// 启动gRPC服务器
	if cfg.GRPCServer.Port > 0 {
		grpcHost := cfg.GRPCServer.Host
		if grpcHost == "" {
			grpcHost = "0.0.0.0"
		}
		grpcAddr = fmt.Sprintf("%s:%d", grpcHost, cfg.GRPCServer.Port)

		logger.Info("正在启动上传服务gRPC服务器...", map[string]interface{}{
			"address": grpcAddr,
		})

		grpcListener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("监听gRPC端口失败", map[string]interface{}{
				"address": grpcAddr,
				"error":   err,
			})
			return
		}

		grpcServer = grpc.NewServer()
		videoService := service.NewVideoPublishService()
		uploadpb.RegisterUploadServiceServer(grpcServer, uploadGrpc.NewUploadGrpcServer(videoService))

		go func() {
			if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				logger.Error("gRPC服务器异常退出", map[string]interface{}{"error": err})
			}
		}()

		logger.Info("上传服务gRPC服务器已启动", map[string]interface{}{
			"address": grpcAddr,
		})

		if cfg.GRPCServiceRegistry.ServiceName != "" && cfg.GRPCServiceRegistry.ServiceID != "" && cfg.GRPCServiceRegistry.Enabled && len(cfg.Etcd.Endpoints) > 0 {
			registerHost := cfg.GRPCServiceRegistry.RegisterHost
			if registerHost == "" {
				registerHost = grpcHost
				if registerHost == "" || registerHost == "0.0.0.0" {
					registerHost = "localhost"
				}
			}
			serviceAddr := fmt.Sprintf("%s:%d", registerHost, cfg.GRPCServer.Port)

			grpcServiceConfig := registry.ServiceConfig{
				ServiceName:     cfg.GRPCServiceRegistry.ServiceName,
				ServiceID:       cfg.GRPCServiceRegistry.ServiceID,
				TTL:             cfg.GRPCServiceRegistry.TTL,
				RefreshInterval: cfg.GRPCServiceRegistry.RefreshInterval,
			}

			grpcRegistry, err = registry.NewServiceRegistry(registryCfg, grpcServiceConfig, serviceAddr)
			if err != nil {
				logger.Fatal("创建gRPC服务注册失败", map[string]interface{}{
					"error": err,
				})
				return
			}
			if err := grpcRegistry.Register(); err != nil {
				logger.Fatal("注册gRPC服务到etcd失败", map[string]interface{}{
					"error": err,
				})
				return
			}
			logger.Info("上传服务gRPC实例已注册到etcd", map[string]interface{}{
				"service": cfg.GRPCServiceRegistry.ServiceName,
				"address": serviceAddr,
			})
		} else if cfg.GRPCServiceRegistry.ServiceName != "" && !cfg.GRPCServiceRegistry.Enabled {
			logger.Info("跳过gRPC服务注册（未开启或未配置etcd）", map[string]interface{}{
				"service": cfg.GRPCServiceRegistry.ServiceName,
			})
		}
	} else {
		logger.Warn("gRPC server port is not configured, skipping gRPC server startup", nil)
	}

	task.StartChunkCleanupTask()
	task.StartMergeTask()

	// 创建Gin引擎
	logger.Info("正在创建HTTP路由...")
	router := gin.Default()

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "upload-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 注册所有路由
	logger.Info("正在注册所有路由...")
	manager.RegisterAllRoutes(router)
	logger.Info("路由注册完成")

	// 启动HTTP服务器
	port := getEnv("PORT", "8082")
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
		"service":    "upload-service",
		"health_url": fmt.Sprintf("http://localhost:%s/health", port),
		"api_url":    fmt.Sprintf("http://localhost:%s/api/v1", port),
	})

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到关闭信号，正在优雅关闭服务器...")

	if grpcRegistry != nil {
		logger.Info("正在注销gRPC服务注册...", map[string]interface{}{
			"service": cfg.GRPCServiceRegistry.ServiceName,
		})
		if err := grpcRegistry.Deregister(); err != nil {
			logger.Error("注销gRPC服务失败", map[string]interface{}{"error": err})
		}
	}

	if grpcServer != nil {
		logger.Info("正在停止gRPC服务器...", map[string]interface{}{"address": grpcAddr})
		grpcServer.GracefulStop()
	}
	if grpcListener != nil {
		_ = grpcListener.Close()
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

	fmt.Println("[SHUTDOWN] 上传服务已安全退出")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
