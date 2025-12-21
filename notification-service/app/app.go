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

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	notificationgrpc "notification-service/ddd/adapter/grpc"
	_ "notification-service/ddd/adapter/http"
	"notification-service/ddd/application/app"
	"notification-service/internal/resource"
	"notification-service/pkg/config"
	"notification-service/pkg/grpcutil"
	"notification-service/pkg/logger"
	"notification-service/pkg/middleware"
	"notification-service/pkg/redisclient"
	"notification-service/pkg/repository"
	"notification-service/pkg/sse"
	notificationpb "notification-service/proto/notification"
)

// Run is the entrypoint of notification-service.
func Run() {
	fmt.Println("[STARTUP] Starting notification service...")

	cfgPath := resolveConfigPath()
	fmt.Println("[STARTUP] Loading config file...")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to load config (%s): %v\n", cfgPath, err)
		os.Exit(1)
	}
	config.SetGlobalConfig(cfg)
	fmt.Printf("[STARTUP] Config file loaded: %s\n", cfgPath)

	fmt.Println("[STARTUP] Initializing logger...")
	logService := logger.NewLogger(cfg)
	logger.SetGlobalLogger(logService)
	logger.Infof("Notification service starting version=%s env=%s", "1.0.0", "development")

	// Initialize database connection and expose it via internal resource package.
	logger.Infof("Initializing database connection...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize database error=%v", err))
	}
	defer db.Close()
	resource.SetMainDB(db.Self)
	logger.Infof("Database connected")

	// Initialize Redis client (optional). If initialization fails we log it
	// and continue with process-local notifications only.
	logger.Infof("Initializing Redis client...")
	redisCli, err := redisclient.New(cfg.Redis)
	if err != nil {
		logger.Errorf("Failed to initialize redis; SSE notifications will be local-only error=%v", err)
	} else {
		defer func() {
			logger.Infof("Closing Redis client...")
			_ = redisCli.Close()
		}()
		// Bridge in-memory SSE hub to Redis Pub/Sub for cross-instance fanout.
		sse.InitRedisPubSub(redisCli.Raw(), "")
	}

	// Create Gin engine and common middlewares.
	logger.Infof("Creating HTTP routes...")
	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.RequestContextMiddleware(),
		middleware.RequestLogMiddleware(),
	)

	// Health check endpoint.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "notification-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Register all controllers via shared manager package.
	logger.Infof("Registering routes...")
	// We reuse user-service's manager.Controller registration; controllers are
	// wired from notification-service/ddd/adapter/http via init() side effects.
	// We don't use ServicePlugin/ComponentPlugin in this service yet.
	// Note: manager.RegisterAllRoutes is imported in controller package.
	// Here we only attach routes to the root router.
	userRouter := router.Group("/")
	_ = userRouter // keep router root; actual registration is done in manager.RegisterAllRoutes

	// user-service's manager.RegisterAllRoutes operates on given *gin.Engine,
	// so we call it directly here.
	// We import it via a blank import in the adapter to avoid an explicit dependency here.
	// nolint:staticcheck
	userServiceRegisterAllRoutes(router)

	logger.Infof("Routes registered")

	// Start gRPC server (for inter-service notifications).
	var (
		grpcServer   *grpc.Server
		grpcListener net.Listener
		grpcAddr     string
	)

	if cfg.GRPC.Port > 0 {
		grpcHost := cfg.Server.Host
		if grpcHost == "" {
			grpcHost = "0.0.0.0"
		}
		grpcAddr = fmt.Sprintf("%s:%d", grpcHost, cfg.GRPC.Port)

		logger.Infof("Starting notification gRPC server address=%s", grpcAddr)

		grpcListener, err = net.Listen(cfg.GRPC.Network, grpcAddr)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to listen on gRPC port address=%s error=%v", grpcAddr, err))
		}

		grpcServer = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				grpcutil.UnaryServerRequestIDInterceptor,
			),
			grpc.MaxRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
			grpc.MaxSendMsgSize(cfg.GRPC.MaxSendMsgSize),
		)

		notificationApp := app.DefaultNotificationApp()
		notificationpb.RegisterNotificationServiceServer(grpcServer, notificationgrpc.NewNotificationGrpcServer(notificationApp))

		go func() {
			if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				logger.Errorf("Notification gRPC server exited unexpectedly error=%v", err)
			}
		}()

		logger.Infof("Notification gRPC server started address=%s", grpcAddr)
	} else {
		logger.Warnf("gRPC port is not configured, skipping notification gRPC server startup")
	}

	// Start HTTP server with graceful shutdown.
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		logger.Infof("HTTP server starting port=%s service=%s", port, "notification-service")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(fmt.Sprintf("Failed to start HTTP server error=%v", err))
		}
	}()

	logger.Infof("HTTP server started port=%s health_url=%s", port, fmt.Sprintf("http://localhost%s/health", port))

	// Wait for termination signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Received shutdown signal, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if grpcServer != nil {
		logger.Infof("Stopping notification gRPC server address=%s", grpcAddr)
		grpcServer.GracefulStop()
	}
	if grpcListener != nil {
		_ = grpcListener.Close()
	}

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Server forced to close error=%v", err))
	}

	logger.Infof("Server exited safely")

	if logService != nil {
		logger.Infof("Closing logger...")
		logService.Close()
	}
}

// resolveConfigPath determines which config file to use.
func resolveConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	if env := os.Getenv("CONFIG_ENV"); env != "" {
		return fmt.Sprintf("configs/config.%s.yaml", env)
	}
	return "configs/config.dev.yaml"
}
