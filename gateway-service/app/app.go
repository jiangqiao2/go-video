package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gateway-service/pkg/auth"
	"gateway-service/pkg/config"
	"gateway-service/pkg/logger"
	"gateway-service/pkg/middleware"
	"gateway-service/pkg/proxy"
)

// Run boots the gateway service.
func Run() {
	fmt.Println("[STARTUP] Bootstrapping gateway-service...")

	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] setup logger: %v\n", err)
		os.Exit(1)
	}
	log.WithFields(logrus.Fields{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
		"mode": cfg.Server.Mode,
	}).Info("gateway configuration loaded")

	gin.SetMode(strings.ToLower(cfg.Server.Mode))
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger(log))
	engine.Use(corsMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "gateway-service",
			"timestamp": time.Now().Unix(),
		})
	})

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "未找到路由",
			"path":    c.Request.URL.Path,
		})
	})

	jwtUtil := auth.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.ExpireTime, cfg.JWT.RefreshExpireTime)
	authenticator := middleware.NewAuthenticator(jwtUtil, log)

	proxyManager, err := proxy.NewManager(cfg.Services, log)
	if err != nil {
		log.WithError(err).Fatal("initialize proxy manager")
		return
	}

	registerRoutes(engine, cfg.Routes, authenticator, proxyManager, log)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Infof("gateway-service listening on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("gateway server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down gateway-service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("graceful shutdown failed")
	} else {
		log.Info("gateway-service stopped gracefully")
	}
}

func registerRoutes(router *gin.Engine, routes []config.RouteConfig, authenticator *middleware.Authenticator, proxyManager *proxy.Manager, log *logrus.Logger) {
	const defaultName = "anonymous-route"
	for idx, route := range routes {
		if route.Name == "" {
			route.Name = fmt.Sprintf("%s-%d", defaultName, idx)
		}

		handler, err := proxyManager.Handler(route)
		if err != nil {
			log.WithError(err).Fatal("register route failed")
			continue
		}

		wrapped := authenticator.Wrap(handler, route.AuthRequired)
		registerRoute(router, route, wrapped)
		log.WithFields(logrus.Fields{
			"name":          route.Name,
			"path_prefix":   route.PathPrefix,
			"methods":       route.Methods,
			"auth_required": route.AuthRequired,
			"target":        route.TargetService,
		}).Info("registered gateway route")
	}
}

func registerRoute(router *gin.Engine, route config.RouteConfig, handler gin.HandlerFunc) {
	methods := route.Methods
	if len(methods) == 0 {
		methods = []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		}
	}

	basePath := normalizePath(route.PathPrefix)
	wildcardPath := buildWildcard(basePath)

	for _, method := range methods {
		router.Handle(method, basePath, handler)
		if wildcardPath != basePath {
			router.Handle(method, wildcardPath, handler)
		}
	}
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	cleaned := strings.TrimRight(path, "/")
	if cleaned == "" {
		return "/"
	}
	return cleaned
}

func buildWildcard(base string) string {
	if base == "/" {
		return "/*proxyPath"
	}
	return fmt.Sprintf("%s/*proxyPath", base)
}

func requestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		entry := log.WithFields(logrus.Fields{
			"status":      status,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"ip":          c.ClientIP(),
			"latency_ms":  latency.Milliseconds(),
			"user_agent":  c.Request.UserAgent(),
		})

		switch {
		case status >= 500:
			entry.Error("request completed")
		case status >= 400:
			entry.Warn("request completed")
		default:
			entry.Info("request completed")
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Requested-With,X-User-UUID,X-User-ID")
		if origin != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

