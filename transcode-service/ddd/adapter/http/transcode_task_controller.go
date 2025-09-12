package http

import (
	"fmt"
	"strconv"
	"transcode-service/ddd/application/app"
	"transcode-service/ddd/application/dto"
	"transcode-service/pkg/restapi"
	"github.com/gin-gonic/gin"
)

// TranscodeTaskController 转码任务控制器
type TranscodeTaskController struct {
	schedulerApp app.SchedulerApp
}

// NewTranscodeTaskController 创建转码任务控制器
func NewTranscodeTaskController(schedulerApp app.SchedulerApp) *TranscodeTaskController {
	return &TranscodeTaskController{
		schedulerApp: schedulerApp,
	}
}

// CreateTask 创建转码任务
func (c *TranscodeTaskController) CreateTask(ctx *gin.Context) {
	var req dto.CreateTranscodeTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	resp, err := c.schedulerApp.CreateTask(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// GetTask 获取任务详情
func (c *TranscodeTaskController) GetTask(ctx *gin.Context) {
	req := dto.GetTaskRequest{
		TaskID: ctx.Param("task_id"),
	}
	
	if req.TaskID == "" {
		restapi.Failed(ctx, fmt.Errorf("task ID is required"))
		return
	}
	
	resp, err := c.schedulerApp.GetTask(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	if resp.Task == nil {
		restapi.Failed(ctx, fmt.Errorf("task not found"))
		return
	}
	
	restapi.Success(ctx, resp)
}

// ListTasks 获取任务列表
func (c *TranscodeTaskController) ListTasks(ctx *gin.Context) {
	req := dto.ListTasksRequest{
		UserID:   ctx.Query("user_id"),
		WorkerID: ctx.Query("worker_id"),
		OrderBy:  ctx.DefaultQuery("order_by", "created_at"),
	}
	
	// 解析状态参数
	if statusStr := ctx.Query("status"); statusStr != "" {
		req.Status = []string{statusStr}
	}
	
	// 解析优先级
	if priorityStr := ctx.Query("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			req.Priority = &priority
		}
	}
	
	// 解析分页参数
	req.Limit, _ = strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	req.OrderDesc, _ = strconv.ParseBool(ctx.DefaultQuery("order_desc", "true"))
	
	resp, err := c.schedulerApp.ListTasks(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// UpdateTaskStatus 更新任务状态
func (c *TranscodeTaskController) UpdateTaskStatus(ctx *gin.Context) {
	var req dto.UpdateTaskStatusRequest
	req.TaskID = ctx.Param("task_id")
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	err := c.schedulerApp.UpdateTaskStatus(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"message": "Task status updated successfully"})
}

// UpdateTaskProgress 更新任务进度
func (c *TranscodeTaskController) UpdateTaskProgress(ctx *gin.Context) {
	var req dto.UpdateTaskProgressRequest
	req.TaskID = ctx.Param("task_id")
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	err := c.schedulerApp.UpdateTaskProgress(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"message": "Task progress updated successfully"})
}

// CancelTask 取消任务
func (c *TranscodeTaskController) CancelTask(ctx *gin.Context) {
	req := dto.CancelTaskRequest{
		TaskID: ctx.Param("task_id"),
	}
	
	if req.TaskID == "" {
		restapi.Failed(ctx, fmt.Errorf("task ID is required"))
		return
	}
	
	err := c.schedulerApp.CancelTask(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"message": "Task cancelled successfully"})
}

// RetryTask 重试任务
func (c *TranscodeTaskController) RetryTask(ctx *gin.Context) {
	req := dto.RetryTaskRequest{
		TaskID: ctx.Param("task_id"),
	}
	
	if req.TaskID == "" {
		restapi.Failed(ctx, fmt.Errorf("task ID is required"))
		return
	}
	
	err := c.schedulerApp.RetryTask(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, gin.H{"message": "Task retry initiated successfully"})
}

// GetTaskStatistics 获取任务统计
func (c *TranscodeTaskController) GetTaskStatistics(ctx *gin.Context) {
	resp, err := c.schedulerApp.GetTaskStatistics(ctx.Request.Context())
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}

// BatchOperation 批量操作任务
func (c *TranscodeTaskController) BatchOperation(ctx *gin.Context) {
	var req dto.BatchOperationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	resp, err := c.schedulerApp.BatchOperation(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	
	restapi.Success(ctx, resp)
}