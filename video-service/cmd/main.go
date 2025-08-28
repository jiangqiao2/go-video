package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin引擎
	router := gin.Default()

	// 视频相关路由（移除上传功能）
	v1 := router.Group("/api/v1/videos")
	{
		v1.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Get video list",
				"service": "video-service",
			})
		})
		v1.GET("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Get video detail",
				"video_id": c.Param("id"),
				"service": "video-service",
			})
		})
		v1.GET("/:id/play", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Get video play url",
				"video_id": c.Param("id"),
				"service": "video-service",
			})
		})
		v1.POST("/:id/like", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Like video",
				"video_id": c.Param("id"),
				"service": "video-service",
			})
		})
		v1.GET("/search", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Search videos",
				"query": c.Query("q"),
				"service": "video-service",
			})
		})
		v1.GET("/recommend", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Get recommended videos",
				"service": "video-service",
			})
		})
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "video-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 启动HTTP服务器
	port := getEnv("PORT", "8083")
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

	fmt.Printf("Video service started on port %s\n", port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down video service...")

	// 5秒超时关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Video service forced to shutdown:", err)
	}

	fmt.Println("Video service exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}