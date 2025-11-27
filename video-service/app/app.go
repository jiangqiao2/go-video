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

	"video-service/pkg/config"
	"video-service/pkg/logger"
	"video-service/pkg/manager"
	"video-service/pkg/redisclient"
	"video-service/pkg/repository"
	"video-service/pkg/utils"

	_ "video-service/ddd/adapter/http"
)

func Run() {
	fmt.Println("[STARTUP] Starting video service...")

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

	manager.MustInitResources()
	defer manager.CloseResources()

	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to init database", map[string]interface{}{"error": err})
		return
	}
	defer db.Close()

	redisCli, err := redisclient.New(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to init redis", map[string]interface{}{"error": err})
		return
	}
	defer func() { _ = redisCli.Close() }()

	jwtUtil := utils.DefaultJWTUtil()

	deps := &manager.Dependencies{DB: db.Self, Config: cfg, JWTUtil: jwtUtil, Redis: redisCli}

	manager.MustInitServices(deps)
	manager.MustInitComponents(deps)

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

		logger.Info("Starting video gRPC server", map[string]interface{}{"address": grpcAddr})

		grpcListener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("Failed to listen on gRPC port", map[string]interface{}{"address": grpcAddr, "error": err})
			return
		}

		grpcServer = grpc.NewServer()
		go func() {
			if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				logger.Error("Video gRPC server exited unexpectedly", map[string]interface{}{"error": err})
			}
		}()

		logger.Info("Video gRPC server started", map[string]interface{}{"address": grpcAddr})
	} else {
		logger.Warn("gRPC server port is not configured, skipping gRPC server startup", nil)
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "video-service",
			"timestamp": time.Now().Unix(),
		})
	})

	manager.RegisterAllRoutes(router)

	port := getEnv("PORT", "8083")
	server := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start HTTP server", map[string]interface{}{"error": err})
		}
	}()

	logger.Info("HTTP server started", map[string]interface{}{
		"port":       port,
		"service":    "video-service",
		"health_url": fmt.Sprintf("http://localhost:%s/health", port),
		"api_url":    fmt.Sprintf("http://localhost:%s/api/v1", port),
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	manager.Shutdown()

	if grpcServer != nil {
		logger.Info("Stopping gRPC server", map[string]interface{}{"address": grpcAddr})
		grpcServer.GracefulStop()
	}
	if grpcListener != nil {
		_ = grpcListener.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	fmt.Println("[SHUTDOWN] Video service exited safely")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
