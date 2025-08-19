package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-video/pkg/config"
	"go-video/pkg/logger"
	"go-video/pkg/manager"
	"go-video/pkg/repository"
	"go-video/pkg/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	// 先使用标准输出确保能看到日志
	fmt.Println("[STARTUP] 开始启动应用程序...")

	// 加载配置
	fmt.Println("[STARTUP] 正在加载配置文件...")
	cfg, err := config.Load("configs/config.dev.yaml")
	if err != nil {
		fmt.Printf("[ERROR] 加载配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[STARTUP] 配置文件加载成功")

	// 初始化日志服务
	fmt.Println("[STARTUP] 正在初始化日志服务...")
	logService := logger.NewLogger(cfg)
	logger.SetGlobalLogger(logService)
	fmt.Println("[STARTUP] 日志服务初始化完成")

	logger.Info("应用程序启动", map[string]interface{}{"version": "1.0.0", "env": "development"})

	// 资源管理器初始化
	logger.Info("正在初始化资源管理器...")
	manager.MustInitResources()
	defer manager.CloseResources()
	logger.Info("资源管理器初始化完成")

	// 初始化数据库
	logger.Info("正在初始化数据库连接...")
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("初始化数据库失败", map[string]interface{}{"error": err})
	}
	defer db.Close()
	logger.Info("数据库连接成功")

	// 初始化JWT工具
	logger.Info("正在初始化JWT工具...")
	jwtUtil := utils.DefaultJWTUtil()
	logger.Info("JWT工具初始化成功")

	// 创建依赖注入容器
	deps := &manager.Dependencies{
		DB:      db.Self,
		Config:  cfg,
		JWTUtil: jwtUtil,
	}

	// 初始化所有服务
	logger.Info("正在初始化所有服务...")
	manager.MustInitServices(deps)
	logger.Info("所有服务初始化完成")

	// 初始化所有组件
	logger.Info("正在初始化所有组件...")
	manager.MustInitComponents(deps)
	logger.Info("所有组件初始化完成")

	// 创建Gin引擎
	logger.Info("正在创建HTTP路由...")
	router := gin.Default()

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// 注册所有路由
	logger.Info("正在注册所有路由...")
	manager.RegisterAllRoutes(router)
	logger.Info("路由注册完成")

	// 启动HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("启动服务器失败", map[string]interface{}{"error": err})
		}
	}()

	logger.Info("HTTP服务器启动成功", map[string]interface{}{
		"port":       cfg.Server.Port,
		"mode":       cfg.Server.Mode,
		"health_url": fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port),
		"api_url":    fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port),
	})

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到关闭信号，正在优雅关闭服务器...")

	// 关闭所有组件
	logger.Info("正在关闭所有组件...")
	manager.Shutdown()
	logger.Info("所有组件已关闭")

	// 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", map[string]interface{}{"error": err})
	}

	logger.Info("服务器已安全退出")

	// 关闭日志服务
	logger.Info("正在关闭日志服务...")
	if logService != nil {
		logService.Close()
	}

	fmt.Println("[SHUTDOWN] 应用程序已安全退出")
}
