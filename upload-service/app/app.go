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
	"upload-service/ddd/domain/service"

	"github.com/gin-gonic/gin"

	uploadpb "upload-service/proto/upload"

	"google.golang.org/grpc"

	uploadGrpc "upload-service/ddd/adapter/grpc"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	"upload-service/pkg/config"
	"upload-service/pkg/kafka"
	"upload-service/pkg/logger"
	"upload-service/pkg/manager"
	"upload-service/pkg/repository"
	"upload-service/pkg/utils"

	_ "upload-service/ddd/adapter/http"

	// 导入资源和模块包以触发init函数
	_ "upload-service/internal/resource"
)

func Run() {
	// 先使用标准输出确保能看到日志
	fmt.Println("[STARTUP] Starting upload service...")

	// 加载配置
	fmt.Println("[STARTUP] Loading config file...")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.dev.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to load config: %v\n", err)
		os.Exit(1)
	}
	// 设置全局配置（必须在资源管理器初始化之前）
	config.SetGlobalConfig(cfg)
	fmt.Println("[STARTUP] Config file loaded")

	// 立即初始化日志服务（确保所有后续组件都能使用正确的日志器）
	fmt.Println("[STARTUP] Initializing logger...")
	logService := logger.NewLogger(cfg)
	logger.SetGlobalLogger(logService)
	fmt.Println("[STARTUP] Logger initialized")

	// 验证日志器配置
	logger.Debug("Logger initialized", map[string]interface{}{
		"level":  cfg.Log.Level,
		"format": cfg.Log.Format,
		"output": cfg.Log.Output,
	})

	logger.Info("Upload service starting", map[string]interface{}{"version": "1.0.0", "env": "development"})

	// 资源管理器初始化
	logger.Info("Initializing resource manager...")
	manager.MustInitResources()
	defer manager.CloseResources()
	logger.Info("Resource manager initialized")

	// 初始化数据库（用于依赖注入）
	logger.Info("Initializing database connection...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", map[string]interface{}{"error": err})
	}
	defer db.Close()
	logger.Info("Database connected")

	// 初始化JWT工具
	logger.Info("Initializing JWT utility...")
	jwtUtil := utils.DefaultJWTUtil()
	logger.Info("JWT utility initialized")

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:      db.Self,
		Config:  cfg,
		JWTUtil: jwtUtil,
		Kafka:   kafka.DefaultClient(),
	}

	// 初始化gRPC客户端（直连/k3s服务名）
	logger.Info("Initializing gRPC clients...")
	clientConfig := grpcClient.ClientConfig{
		Timeout:        cfg.GRPC.Timeout,
		MaxRecvMsgSize: cfg.GRPC.MaxRecvMsgSize,
		MaxSendMsgSize: cfg.GRPC.MaxSendMsgSize,
		RetryTimes:     cfg.GRPC.RetryTimes,
	}
	userServiceClient, err := grpcClient.NewUserServiceClient(clientConfig)
	if err != nil {
		logger.Fatal("Failed to create user gRPC client", map[string]interface{}{"error": err})
		return
	}
	deps.UserServiceClient = userServiceClient
	logger.Info("gRPC clients initialized")

	// 初始化所有服务（在gRPC客户端初始化之后）
	logger.Info("Initializing services...")
	manager.MustInitServices(deps)
	logger.Info("All services initialized")

	// 初始化所有组件
	logger.Info("Initializing components...")
	manager.MustInitComponents(deps)
	logger.Info("All components initialized")

	// 启动gRPC服务器（保留RPC接口，同时结果通过Kafka回传）
	var (
		grpcListener net.Listener
		grpcServer   *grpc.Server
		grpcAddr     string
	)

	if cfg.GRPCServer.Port > 0 {
		grpcHost := cfg.GRPCServer.Host
		if grpcHost == "" {
			grpcHost = "0.0.0.0"
		}
		grpcAddr = fmt.Sprintf("%s:%d", grpcHost, cfg.GRPCServer.Port)

		logger.Info("Starting upload gRPC server...", map[string]interface{}{
			"address": grpcAddr,
		})

		grpcListener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("Failed to listen on gRPC port", map[string]interface{}{
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
				logger.Error("Upload gRPC server exited unexpectedly", map[string]interface{}{"error": err})
			}
		}()

		logger.Info("Upload gRPC server started", map[string]interface{}{
			"address": grpcAddr,
		})
	} else {
		logger.Warn("gRPC server port is not configured, skipping gRPC server startup", nil)
	}

	// 启动后台任务
	task.StartChunkCleanupTask()
	task.StartMergeTask()

	// 创建Gin引擎
	logger.Info("Creating HTTP routes...")
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
	logger.Info("Registering routes...")
	manager.RegisterAllRoutes(router)
	logger.Info("Routes registered")

	// 启动HTTP服务器
	port := getEnv("PORT", "8082")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start HTTP server", map[string]interface{}{"error": err})
		}
	}()

	logger.Info("HTTP server started", map[string]interface{}{
		"port":       port,
		"service":    "upload-service",
		"health_url": fmt.Sprintf("http://localhost:%s/health", port),
		"api_url":    fmt.Sprintf("http://localhost:%s/api/v1", port),
	})

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Received shutdown signal, shutting down server...")

	if grpcServer != nil {
		logger.Info("Stopping gRPC server...", map[string]interface{}{"address": grpcAddr})
		grpcServer.GracefulStop()
	}
	if grpcListener != nil {
		_ = grpcListener.Close()
	}

	// 关闭所有组件
	logger.Info("Shutting down components...")
	manager.Shutdown()
	logger.Info("Components closed")

	// 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to close", map[string]interface{}{"error": err})
	}

	logger.Info("Server exited safely")

	// 关闭日志服务
	logger.Info("Closing logger...")
	if logService != nil {
		logService.Close()
	}

	fmt.Println("[SHUTDOWN] Upload service exited safely")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
