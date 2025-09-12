package http

import (
	"fmt"
	"strconv"
	"time"
	"transcode-service/ddd/application/app"
	"transcode-service/ddd/application/dto"
	"transcode-service/pkg/restapi"
	"github.com/gin-gonic/gin"
)

// WorkerController Worker控制器
type WorkerController struct {
	workerApp app.WorkerApp
}

// NewWorkerController 创建Worker控制器
func NewWorkerController(workerApp app.WorkerApp) *WorkerController {
	return &WorkerController{
		workerApp: workerApp,
	}
}

// RegisterWorker 注册Worker
func (c *WorkerController) RegisterWorker(ctx *gin.Context) {
	var req dto.RegisterWorkerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	resp, err := c.workerApp.RegisterWorker(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// GetWorker 获取Worker详情
func (c *WorkerController) GetWorker(ctx *gin.Context) {
	req := dto.GetWorkerRequest{
		WorkerID: ctx.Param("worker_id"),
	}
	
	if req.WorkerID == "" {
		restapi.Failed(ctx, fmt.Errorf("worker ID is required"))
		return
	}
	
	resp, err := c.workerApp.GetWorker(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	if resp.Worker == nil {
		restapi.Failed(ctx, fmt.Errorf("worker not found"))
		return
	}
	
	restapi.Success(ctx, resp)
}

// ListWorkers 获取Worker列表
func (c *WorkerController) ListWorkers(ctx *gin.Context) {
	req := dto.ListWorkersRequest{
		OrderBy: ctx.DefaultQuery("order_by", "registered_at"),
	}
	
	// 解析状态参数
	if statusStr := ctx.Query("status"); statusStr != "" {
		req.Status = []string{statusStr}
	}
	
	// 解析健康状态
	if healthyStr := ctx.Query("healthy"); healthyStr != "" {
		if healthy, err := strconv.ParseBool(healthyStr); err == nil {
			req.Healthy = &healthy
		}
	}
	
	// 解析分页参数
	req.Limit, _ = strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	req.OrderDesc, _ = strconv.ParseBool(ctx.DefaultQuery("order_desc", "true"))
	
	resp, err := c.workerApp.ListWorkers(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// UpdateWorkerStatus 更新Worker状态
func (c *WorkerController) UpdateWorkerStatus(ctx *gin.Context) {
	var req dto.UpdateWorkerStatusRequest
	req.WorkerID = ctx.Param("worker_id")
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	err := c.workerApp.UpdateWorkerStatus(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"message": "Worker status updated successfully"})
}

// ProcessHeartbeat 处理Worker心跳
func (c *WorkerController) ProcessHeartbeat(ctx *gin.Context) {
	var req dto.WorkerHeartbeatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	resp, err := c.workerApp.ProcessHeartbeat(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// GetWorkerStatistics 获取Worker统计
func (c *WorkerController) GetWorkerStatistics(ctx *gin.Context) {
	resp, err := c.workerApp.GetWorkerStatistics(ctx.Request.Context())
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// DeleteWorker 删除Worker
func (c *WorkerController) DeleteWorker(ctx *gin.Context) {
	req := dto.DeleteWorkerRequest{
		WorkerID: ctx.Param("worker_id"),
	}
	
	if req.WorkerID == "" {
		restapi.Failed(ctx, fmt.Errorf("worker ID is required"))
		return
	}
	
	// 解析强制删除参数
	req.Force, _ = strconv.ParseBool(ctx.DefaultQuery("force", "false"))
	
	err := c.workerApp.DeleteWorker(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"message": "Worker deleted successfully"})
}

// GetWorkerTasks 获取Worker任务列表
func (c *WorkerController) GetWorkerTasks(ctx *gin.Context) {
	req := dto.WorkerTasksRequest{
		WorkerID: ctx.Param("worker_id"),
	}
	
	if req.WorkerID == "" {
		restapi.Failed(ctx, fmt.Errorf("worker ID is required"))
		return
	}
	
	// 解析状态参数
	if statusStr := ctx.Query("status"); statusStr != "" {
		req.Status = []string{statusStr}
	}
	
	// 解析分页参数
	req.Limit, _ = strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	
	resp, err := c.workerApp.GetWorkerTasks(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// BatchWorkerOperation 批量Worker操作
func (c *WorkerController) BatchWorkerOperation(ctx *gin.Context) {
	var req dto.BatchWorkerOperationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	resp, err := c.workerApp.BatchWorkerOperation(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// CheckWorkerHealth 检查Worker健康状态
func (c *WorkerController) CheckWorkerHealth(ctx *gin.Context) {
	workerID := ctx.Param("worker_id")
	if workerID == "" {
		restapi.Failed(ctx, fmt.Errorf("worker ID is required"))
		return
	}
	
	resp, err := c.workerApp.CheckWorkerHealth(ctx.Request.Context(), workerID)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// CleanupUnhealthyWorkers 清理不健康的Worker
func (c *WorkerController) CleanupUnhealthyWorkers(ctx *gin.Context) {
	timeoutSeconds, _ := strconv.Atoi(ctx.DefaultQuery("timeout", "60"))
	timeout := time.Duration(timeoutSeconds) * time.Second
	
	cleanedCount, err := c.workerApp.CleanupUnhealthyWorkers(ctx.Request.Context(), timeout)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"cleaned_count": cleanedCount})
}