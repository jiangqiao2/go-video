package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"user-service/pkg/config"
	grpcServer "user-service/pkg/grpc"
	"user-service/pkg/logger"
	"user-service/pkg/manager"
	"user-service/pkg/repository"
	"user-service/pkg/utils"
	pb "user-service/proto/user"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	_ "user-service/ddd/adapter/http"
	app "user-service/ddd/application/app"
	_ "user-service/ddd/infrastructure/cache"
)

func Run() {
	// 先使用标准输出确保能看到日志
	fmt.Println("[STARTUP] Starting user service...")

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
	logger.Debug(fmt.Sprintf("Logger initialized level=%s format=%s output=%s", cfg.Log.Level, cfg.Log.Format, cfg.Log.Output))

	logger.Infof("User service starting version=%s env=%s", "1.0.0", "development")

	// 资源管理器初始化
	logger.Infof("Initializing resource manager...")
	manager.MustInitResources()
	defer manager.CloseResources()
	logger.Infof("Resource manager initialized")

	// Revocation store is initialized by infrastructure component plugin

	// 初始化数据库（用于依赖注入）
	logger.Infof("Initializing database connection...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize database error=%v", err))
	}
	defer db.Close()
	logger.Infof("Database connected")

	// 初始化JWT工具
	logger.Infof("Initializing JWT utility...")
	jwtUtil := utils.DefaultJWTUtil()
	logger.Infof("JWT utility initialized")

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:      db.Self,
		Config:  cfg,
		JWTUtil: jwtUtil,
	}

	// 初始化所有服务
	logger.Infof("Initializing services...")
	manager.MustInitServices(deps)
	logger.Infof("All services initialized")

	// 初始化所有组件
	logger.Infof("Initializing components...")
	manager.MustInitComponents(deps)
	logger.Infof("All components initialized")

	// 启动gRPC服务
	logger.Infof("Starting gRPC server... port=%d", cfg.GRPC.Port)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to listen on gRPC port: %v", err))
		return
	}

	grpcSrv := grpc.NewServer()
	userApp := app.DefaultUserApp()
	userServiceServer := grpcServer.NewUserServiceServer(userApp)
	// 注册gRPC服务
	pb.RegisterUserServiceServer(grpcSrv, userServiceServer)

	go func() {
		logger.Infof("gRPC server started, listening on port: %d", cfg.GRPC.Port)
		if err := grpcSrv.Serve(grpcListener); err != nil {
			logger.Errorf("gRPC server failed to start: %v", err)
		}
	}()

	// 创建Gin引擎
	logger.Infof("Creating HTTP routes...")
	router := gin.Default()

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "user-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 注册所有路由
	logger.Infof("Registering routes...")
	manager.RegisterAllRoutes(router)
	logger.Infof("Routes registered")

	// 启动HTTP服务器
	port := getEnv("PORT", "8081")
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

	logger.Infof("HTTP server started port=%s service=%s health_url=%s api_url=%s", port, "user-service", fmt.Sprintf("http://localhost:%s/health", port), fmt.Sprintf("http://localhost:%s/api/v1", port))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Received shutdown signal, shutting down server...")

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

	fmt.Println("[SHUTDOWN] User service exited safely")
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
