package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	transcodeGrpc "transcode-service/ddd/adapter/grpc"
	app "transcode-service/ddd/application/app"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/repository"
	"transcode-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	_ "transcode-service/ddd/adapter/component"
	_ "transcode-service/ddd/adapter/http"
	_ "transcode-service/ddd/infrastructure/worker"

	// 导入资源和模块包以触发init函数
	_ "transcode-service/internal/resource"
	transcodepb "transcode-service/proto/transcode"
)

func Run() {
	// 先使用标准输出确保能看到日志
	fmt.Println("[STARTUP] Starting transcode service...")

	// 加载配置
	fmt.Println("[STARTUP] Loading config file...")
	cfgPath := resolveConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to load config (%s): %v\n", cfgPath, err)
		os.Exit(1)
	}
	// 设置全局配置（必须在资源管理器初始化之前）
	config.SetGlobalConfig(cfg)
	fmt.Printf("[STARTUP] Config file loaded: %s\n", cfgPath)

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

	logger.Info("Transcode service starting", map[string]interface{}{"version": "1.0.0", "env": "development"})

	// 检查 FFmpeg 是否可用，直接在启动阶段失败
	ffmpegBin := cfg.Transcode.FFmpeg.BinaryPath
	if strings.TrimSpace(ffmpegBin) == "" {
		ffmpegBin = "ffmpeg"
	}
	if _, err := exec.LookPath(ffmpegBin); err != nil {
		logger.Fatal("FFmpeg binary not found, please install or set transcode.ffmpeg.binary_path", map[string]interface{}{
			"binary": ffmpegBin,
			"error":  err.Error(),
		})
	}

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

	// 初始化转码相关组件
	logger.Info("Initializing transcode components...")

	// 初始化应用服务
	transcodeAppService := app.DefaultTranscodeApp()
	logger.Info("Transcode components initialized")

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:                  db.Self,
		Config:              cfg,
		JWTUtil:             jwtUtil,
		TranscodeAppService: transcodeAppService,
	}

	// 初始化所有服务
	logger.Info("Initializing services...")
	manager.MustInitServices(deps)
	logger.Info("All services initialized")

	// 初始化所有组件
	logger.Info("Initializing components...")
	manager.MustInitComponents(deps)
	logger.Info("All components initialized")

	// 启动gRPC服务（保留RPC接口，同时支持Kafka触发）
	logger.Info("Starting gRPC server...", map[string]interface{}{"address": fmt.Sprintf("%s:%d", cfg.GRPCServer.Host, cfg.GRPCServer.Port)})
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPCServer.Host, cfg.GRPCServer.Port)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("Failed to listen on gRPC port", map[string]interface{}{
			"address": grpcAddr,
			"error":   err,
		})
	}

	grpcServer := grpc.NewServer()
	transcodepb.RegisterTranscodeServiceServer(
		grpcServer,
		transcodeGrpc.NewTranscodeGrpcServer(transcodeAppService),
	)

	go func() {
		logger.Info("gRPC server started", map[string]interface{}{
			"address": grpcAddr,
			"service": "transcode-service",
		})
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("gRPC server encountered an error", map[string]interface{}{
				"error": err,
			})
		}
	}()

	// 创建Gin引擎
	logger.Info("Creating HTTP routes...")
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
	logger.Info("Registering routes...")
	manager.RegisterAllRoutes(router)
	logger.Info("Routes registered")

	// 启动HTTP服务器
	port := getEnv("PORT", "8083")
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
		"service":    "transcode-service",
		"health_url": fmt.Sprintf("http://localhost:%s/health", port),
		"api_url":    fmt.Sprintf("http://localhost:%s/api/v1", port),
	})

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Received shutdown signal, shutting down server...")

	logger.Info("Stopping gRPC server...", map[string]interface{}{"address": grpcAddr})
	grpcServer.GracefulStop()

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

	fmt.Println("[SHUTDOWN] Transcode service exited safely")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// resolveConfigPath 根据环境选择配置文件，支持CONFIG_PATH覆盖、CONFIG_ENV区分环境
func resolveConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	env := strings.ToLower(strings.TrimSpace(os.Getenv("CONFIG_ENV")))
	if env == "" {
		env = "dev"
	}

	switch env {
	case "prod", "production":
		return "configs/config_prod.yaml"
	case "dev", "development":
		return "configs/config.dev.yaml"
	default:
		return fmt.Sprintf("configs/config.%s.yaml", env)
	}
}
