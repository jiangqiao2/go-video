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
	"upload-service/pkg/grpcutil"
	"upload-service/pkg/kafka"
	"upload-service/pkg/logger"
	"upload-service/pkg/manager"
	"upload-service/pkg/middleware"
	"upload-service/pkg/observability"
	"upload-service/pkg/repository"

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

	logger.Infof("Upload service starting version=%s env=%s", "1.0.0", "development")

	rootCtx := context.Background()
	traceCtx, cancelTrace := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancelTrace()
	tracerCfg := observability.DefaultTracerConfig("upload-service")
	shutdownTracer, err := observability.InitTracer(traceCtx, tracerCfg)
	if err != nil {
		logger.Warnf("Failed to initialise tracer provider error=%v", err)
	} else if shutdownTracer != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownTracer(ctx); err != nil {
				logger.Warnf("Failed to shut down tracer provider error=%v", err)
			}
		}()
		logger.Infof("Tracing initialised endpoint=%s sample_ratio=%.2f", tracerCfg.Endpoint, tracerCfg.SampleRatio)
	}

	// 资源管理器初始化
	logger.Infof("Initializing resource manager...")
	manager.MustInitResources()
	defer manager.CloseResources()
	logger.Infof("Resource manager initialized")

	// 初始化数据库（用于依赖注入）
	logger.Infof("Initializing database connection...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize database error=%v", err))
	}
	defer db.Close()
	logger.Infof("Database connected")

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:     db.Self,
		Config: cfg,
		Kafka:  kafka.DefaultClient(),
	}

	// 初始化gRPC客户端（直连/k3s服务名）
	logger.Infof("Initializing gRPC clients...")
	clientConfig := grpcClient.ClientConfig{
		Timeout:        cfg.GRPC.Timeout,
		MaxRecvMsgSize: cfg.GRPC.MaxRecvMsgSize,
		MaxSendMsgSize: cfg.GRPC.MaxSendMsgSize,
		RetryTimes:     cfg.GRPC.RetryTimes,
	}
	userServiceClient, err := grpcClient.NewUserServiceClient(clientConfig)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create user gRPC client error=%v", err))
		return
	}
	deps.UserServiceClient = userServiceClient
	logger.Infof("gRPC clients initialized")

	// 初始化所有服务（在gRPC客户端初始化之后）
	logger.Infof("Initializing services...")
	manager.MustInitServices(deps)
	logger.Infof("All services initialized")

	// 初始化所有组件
	logger.Infof("Initializing components...")
	manager.MustInitComponents(deps)
	logger.Infof("All components initialized")

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

		logger.Infof("Starting upload gRPC server... address=%s", grpcAddr)

		grpcListener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to listen on gRPC port address=%s error=%v", grpcAddr, err))
			return
		}

		grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
			grpcutil.UnaryServerRequestIDInterceptor,
			observability.GRPCServerTracingInterceptor("upload-service"),
		))
		videoService := service.NewVideoPublishService()
		uploadpb.RegisterUploadServiceServer(grpcServer, uploadGrpc.NewUploadGrpcServer(videoService))

		go func() {
			if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				logger.Errorf("Upload gRPC server exited unexpectedly error=%v", err)
			}
		}()

		logger.Infof("Upload gRPC server started address=%s", grpcAddr)
	} else {
		logger.Warnf("gRPC server port is not configured, skipping gRPC server startup")
	}

	// 启动后台任务
	task.StartChunkCleanupTask()
	task.StartMergeTask()

	// 创建Gin引擎
	logger.Infof("Creating HTTP routes...")
	router := gin.New()
	router.Use(
		gin.Recovery(),
		observability.HTTPTraceMiddleware("upload-service"),
		observability.HTTPMetricsMiddleware("upload-service"),
		middleware.RequestContextMiddleware(),
		middleware.RequestLogMiddleware(),
	)

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "upload-service",
			"timestamp": time.Now().Unix(),
		})
	})
	router.GET("/metrics", gin.WrapH(observability.MetricsHandler()))

	// 注册所有路由
	logger.Infof("Registering routes...")
	manager.RegisterAllRoutes(router)
	logger.Infof("Routes registered")

	// 启动HTTP服务器
	port := getEnv("PORT", "8082")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(fmt.Sprintf("Failed to start HTTP server error=%v", err))
		}
	}()

	logger.Infof("HTTP server started port=%s service=%s health_url=%s api_url=%s", port, "upload-service", fmt.Sprintf("http://localhost:%s/health", port), fmt.Sprintf("http://localhost:%s/api/v1", port))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Received shutdown signal, shutting down server...")

	if grpcServer != nil {
		logger.Infof("Stopping gRPC server... address=%s", grpcAddr)
		grpcServer.GracefulStop()
	}
	if grpcListener != nil {
		_ = grpcListener.Close()
	}

	// 关闭所有组件
	logger.Infof("Shutting down components...")
	manager.Shutdown()
	logger.Infof("Components closed")

	// 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Server forced to close error=%v", err))
	}

	logger.Infof("Server exited safely")

	// 关闭日志服务
	logger.Infof("Closing logger...")
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
