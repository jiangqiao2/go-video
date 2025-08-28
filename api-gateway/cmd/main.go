package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin引擎
	router := gin.Default()

	// 服务地址配置
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8081")
	uploadServiceURL := getEnv("UPLOAD_SERVICE_URL", "http://localhost:8082")
	videoServiceURL := getEnv("VIDEO_SERVICE_URL", "http://localhost:8083")

	// 创建反向代理
	userProxy := createProxy(userServiceURL)
	uploadProxy := createProxy(uploadServiceURL)
	videoProxy := createProxy(videoServiceURL)

	// 路由转发配置
	v1 := router.Group("/api/v1")
	{
		// 用户服务路由
		userGroup := v1.Group("/users")
		userGroup.Any("/*path", gin.WrapH(userProxy))

		// 认证服务路由
		authGroup := v1.Group("/auth")
		authGroup.Any("/*path", gin.WrapH(userProxy))

		// 上传服务路由
		uploadGroup := v1.Group("/upload")
		uploadGroup.Any("/*path", gin.WrapH(uploadProxy))

		// 视频服务路由
		videoGroup := v1.Group("/videos")
		videoGroup.Any("/*path", gin.WrapH(videoProxy))
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "api-gateway",
			"timestamp": time.Now().Unix(),
			"services": map[string]string{
				"user": userServiceURL,
				"upload": uploadServiceURL,
				"video": videoServiceURL,
			},
		})
	})

	// 启动HTTP服务器
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Printf("API Gateway started on port %s\n", port)
	fmt.Printf("Proxying to services:\n")
	fmt.Printf("  User Service: %s\n", userServiceURL)
	fmt.Printf("  Upload Service: %s\n", uploadServiceURL)
	fmt.Printf("  Video Service: %s\n", videoServiceURL)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down API Gateway...")

	// 5秒超时关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("API Gateway forced to shutdown:", err)
	}

	fmt.Println("API Gateway exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createProxy(targetURL string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director函数来修改请求
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// 移除API版本前缀，让后端服务处理干净的路径
		// 例如: /api/v1/users/profile -> /api/v1/users/profile
		req.Host = target.Host
	}

	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", r.URL.Path, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"Service temporarily unavailable","code":502}`))
	}

	return proxy
}