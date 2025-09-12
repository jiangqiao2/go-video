package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	httpAdapter "transcode-service/ddd/adapter/http"
	"transcode-service/ddd/application/app"
	"transcode-service/ddd/domain/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志
	// log.SetFormatter(&log.JSONFormatter{})
	// log.SetLevel(log.InfoLevel)
	
	// log.Info("Starting Transcode Scheduler Service...")
	
	// 初始化资源（简化实现，实际项目中需要实现resource包）
	ctx := context.Background()
	// resources, err := resource.InitResources(ctx)
	// if err != nil {
	//	log.WithError(err).Fatal("Failed to initialize resources")
	// }
	// defer resources.Close()
	
	// 初始化仓储（简化实现）
	// taskRepo := persistence.NewTranscodeTaskRepository(resources.DB)
	// workerRepo := persistence.NewWorkerRepository(resources.DB)
	
	// 初始化领域服务（简化实现）
	// taskService := service.NewTranscodeTaskService(taskRepo)
	// workerService := service.NewWorkerService(workerRepo)
	var taskService service.TranscodeTaskService // 简化实现
	var workerService service.WorkerService // 简化实现
	
	// 初始化应用服务
	schedulerApp := app.NewSchedulerApp(taskService, workerService)
	workerApp := app.NewWorkerApp(workerService)
	
	// 初始化HTTP服务器
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	
	// 设置路由
	router := httpAdapter.NewRouter(schedulerApp, workerApp)
	router.SetupMiddleware(engine)
	router.SetupRoutes(engine)
	
	// 启动HTTP服务器
	server := &http.Server{
		Addr:    ":8082",
		Handler: engine,
	}
	
	// 启动调度器后台任务
	go startSchedulerTasks(ctx, schedulerApp)
	
	// 启动服务器
	go func() {
		// log.Info("Starting HTTP server on :8082")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// log.WithError(err).Fatal("Failed to start HTTP server")
			panic(err)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	// log.Info("Shutting down server...")
	
	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		// log.WithError(err).Error("Server forced to shutdown")
	}
	
	// log.Info("Server exited")
}

// startSchedulerTasks 启动调度器后台任务
func startSchedulerTasks(ctx context.Context, schedulerApp app.SchedulerApp) {
	// 任务分配定时器
	assignTicker := time.NewTicker(5 * time.Second)
	defer assignTicker.Stop()
	
	// 清理过期任务定时器
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()
	
	// 处理重试任务定时器
	retryTicker := time.NewTicker(1 * time.Minute)
	defer retryTicker.Stop()
	
	// log.Info("Scheduler background tasks started")
	
	for {
		select {
		case <-ctx.Done():
			// log.Info("Scheduler background tasks stopped")
			return
			
		case <-assignTicker.C:
			// 分配任务给Worker
			if err := schedulerApp.AssignTasks(ctx); err != nil {
				// log.WithError(err).Error("Failed to assign tasks")
			}
			
		case <-cleanupTicker.C:
			// 清理过期任务
			if cleanedCount, err := schedulerApp.CleanupExpiredTasks(ctx); err != nil {
				// log.WithError(err).Error("Failed to cleanup expired tasks")
			} else if cleanedCount > 0 {
				// log.WithField("cleaned_count", cleanedCount).Info("Cleaned up expired tasks")
			}
			
		case <-retryTicker.C:
			// 处理重试任务
			if processedCount, err := schedulerApp.ProcessRetryTasks(ctx); err != nil {
				// log.WithError(err).Error("Failed to process retry tasks")
			} else if processedCount > 0 {
				// log.WithField("processed_count", processedCount).Info("Processed retry tasks")
			}
		}
	}
}