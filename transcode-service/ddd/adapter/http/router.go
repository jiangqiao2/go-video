package http

import (
	"transcode-service/ddd/application/app"
	"github.com/gin-gonic/gin"
)

// Router 路由配置
type Router struct {
	schedulerApp app.SchedulerApp
	workerApp    app.WorkerApp
}

// NewRouter 创建路由配置
func NewRouter(schedulerApp app.SchedulerApp, workerApp app.WorkerApp) *Router {
	return &Router{
		schedulerApp: schedulerApp,
		workerApp:    workerApp,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// 创建控制器
	taskController := NewTranscodeTaskController(r.schedulerApp)
	workerController := NewWorkerController(r.workerApp)
	
	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		// 转码任务相关路由
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskController.CreateTask)                    // 创建任务
			tasks.GET("", taskController.ListTasks)                     // 获取任务列表
			tasks.GET("/statistics", taskController.GetTaskStatistics) // 获取任务统计
			tasks.POST("/batch", taskController.BatchOperation)        // 批量操作任务
			
			tasks.GET("/:task_id", taskController.GetTask)                      // 获取任务详情
			tasks.PUT("/:task_id/status", taskController.UpdateTaskStatus)       // 更新任务状态
			tasks.PUT("/:task_id/progress", taskController.UpdateTaskProgress)   // 更新任务进度
			tasks.POST("/:task_id/cancel", taskController.CancelTask)            // 取消任务
			tasks.POST("/:task_id/retry", taskController.RetryTask)              // 重试任务
		}
		
		// Worker相关路由
		workers := v1.Group("/workers")
		{
			workers.POST("", workerController.RegisterWorker)                    // 注册Worker
			workers.GET("", workerController.ListWorkers)                        // 获取Worker列表
			workers.GET("/statistics", workerController.GetWorkerStatistics)     // 获取Worker统计
			workers.POST("/heartbeat", workerController.ProcessHeartbeat)        // 处理心跳
			workers.POST("/batch", workerController.BatchWorkerOperation)        // 批量操作Worker
			workers.POST("/cleanup", workerController.CleanupUnhealthyWorkers)   // 清理不健康Worker
			
			workers.GET("/:worker_id", workerController.GetWorker)                    // 获取Worker详情
			workers.PUT("/:worker_id/status", workerController.UpdateWorkerStatus)    // 更新Worker状态
			workers.DELETE("/:worker_id", workerController.DeleteWorker)              // 删除Worker
			workers.GET("/:worker_id/health", workerController.CheckWorkerHealth)     // 检查Worker健康
			workers.GET("/:worker_id/tasks", workerController.GetWorkerTasks)         // 获取Worker任务
		}
	}
	
	// 内部API路由组（用于Worker调用）
	internal := engine.Group("/internal/v1")
	{
		// Worker内部接口
		internal.POST("/workers/heartbeat", workerController.ProcessHeartbeat)     // Worker心跳上报
		internal.PUT("/tasks/:task_id/status", taskController.UpdateTaskStatus)    // Worker更新任务状态
		internal.PUT("/tasks/:task_id/progress", taskController.UpdateTaskProgress) // Worker更新任务进度
	}
	
	// 健康检查路由
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "transcode-service",
			"version": "1.0.0",
		})
	})
	
	// 根路径重定向
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Transcode Service API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
		})
	})
}

// SetupMiddleware 设置中间件
func (r *Router) SetupMiddleware(engine *gin.Engine) {
	// CORS中间件
	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// 请求日志中间件
	engine.Use(gin.Logger())
	
	// 恢复中间件
	engine.Use(gin.Recovery())
	
	// 请求限制中间件（可选）
	// engine.Use(ratelimit.RateLimitMiddleware())
	
	// 认证中间件（可选）
	// engine.Use(auth.AuthMiddleware())
}