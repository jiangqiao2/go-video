package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	videoGrpc "video-service/ddd/adapter/grpc"

	videopb "video-service/proto/video"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"video-service/pkg/config"
	"video-service/pkg/grpcutil"
	"video-service/pkg/logger"
	"video-service/pkg/manager"
	"video-service/pkg/middleware"
	"video-service/pkg/redisclient"
	"video-service/pkg/repository"

	_ "video-service/ddd/adapter/http"
)

func Run() {
	fmt.Println("[STARTUP] Starting video service...")
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
	config.SetGlobalConfig(cfg)
	logService := logger.NewLogger(cfg)
	logger.SetGlobalLogger(logService)
	logger.Infof("Config file loaded")
	fmt.Println("[STARTUP] Initializing logger...")
	logger.Debug(fmt.Sprintf("Logger initialized level=%s format=%s output=%s", cfg.Log.Level, cfg.Log.Format, cfg.Log.Output))
	logger.Infof("Video service starting version=%s env=%s", "1.0.0", "development")

	logger.Infof("Initializing resource manager...")
	manager.MustInitResources()
	defer manager.CloseResources()
	logger.Infof("Resource manager initialized")

	logger.Infof("Initializing database connection...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to init database error=%v", err))
		return
	}
	defer db.Close()
	logger.Infof("Database connected")

	logger.Infof("Initializing Redis client...")
	redisCli, err := redisclient.New(cfg.Redis)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to init redis error=%v", err))
		return
	}
	defer func() { _ = redisCli.Close() }()
	logger.Infof("Redis client initialized")

	deps := &manager.Dependencies{DB: db.Self, Config: cfg, Redis: redisCli}

	logger.Infof("Initializing services...")
	manager.MustInitServices(deps)
	logger.Infof("All services initialized")
	logger.Infof("Initializing components...")
	manager.MustInitComponents(deps)
	logger.Infof("All components initialized")

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

		logger.Infof("Starting video gRPC server address=%s", grpcAddr)

		grpcListener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to listen on gRPC port address=%s error=%v", grpcAddr, err))
			return
		}

		grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(grpcutil.UnaryServerRequestIDInterceptor))
		videopb.RegisterVideoServiceServer(grpcServer, &videoGrpc.VideoGRPCServer{})
		go func() {
			if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				logger.Errorf("Video gRPC server exited unexpectedly error=%v", err)
			}
		}()

		logger.Infof("Video gRPC server started address=%s", grpcAddr)
	} else {
		logger.Warnf("gRPC server port is not configured, skipping gRPC server startup")
	}

	logger.Infof("Creating HTTP routes...")
	router := gin.Default()
	router.Use(middleware.RequestContextMiddleware(), middleware.RequestLogMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "video-service",
			"timestamp": time.Now().Unix(),
		})
	})

	logger.Infof("Registering routes...")
	manager.RegisterAllRoutes(router)
	logger.Infof("Routes registered")

	p := cfg.Server.Port
	port := strconv.Itoa(p)
	server := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Failed to start HTTP server error=%v", err)
		}
	}()

	logger.Infof("HTTP server started port=%s service=%s health_url=%s api_url=%s", port, "video-service", fmt.Sprintf("http://localhost:%s/health", port), fmt.Sprintf("http://localhost:%s/api/v1", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Shutting down components...")
	manager.Shutdown()
	logger.Infof("Components closed")

	if grpcServer != nil {
		logger.Infof("Stopping gRPC server address=%s", grpcAddr)
		grpcServer.GracefulStop()
	}
	if grpcListener != nil {
		_ = grpcListener.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	if logService != nil {
		logger.Infof("Closing logger...")
		_ = logService.Close()
	}
	fmt.Println("[SHUTDOWN] Video service exited safely")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
