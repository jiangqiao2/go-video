package http

import (
	"github.com/gin-gonic/gin"
	"sync"
	"transcode-service/ddd/application/app"
	"transcode-service/ddd/application/cqe"
	"transcode-service/pkg/assert"
	"transcode-service/pkg/errno"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/restapi"
)

var (
	transcodeControllerOnce      sync.Once
	singletonTranscodeController TranscodeController
)

type TranscodeControllerPlugin struct {
}

func (p *TranscodeControllerPlugin) Name() string {
	return "transcodeControllerPlugin"
}
func (p *TranscodeControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	transcodeControllerOnce.Do(func() {
		singletonTranscodeController = &transcodeControllerImpl{
			transcodeApp: app.DefaultTranscodeApp(),
		}
	})
	assert.NotNil(singletonTranscodeController)
	return singletonTranscodeController
}

func init() {
	manager.RegisterControllerPlugin(&TranscodeControllerPlugin{})
}

type TranscodeController interface {
	manager.Controller
}

type transcodeControllerImpl struct {
	manager.Controller
	transcodeApp app.TranscodeApp
}

// RegisterOpenApi 注册开放API
func (t *transcodeControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	// 开放API实现
	v1 := router.Group("v1/transcode")
	{
		v1.POST("/tasks", t.CreateTranscodeTask)
		v1.GET("/tasks/:task_uuid", t.GetTranscodeTask)
		v1.GET("/tasks", t.ListTranscodeTasks)
		v1.PUT("/tasks/:task_uuid/status", t.UpdateTranscodeTaskStatus)
		v1.DELETE("/tasks/:task_uuid", t.CancelTranscodeTask)
		v1.GET("/tasks/:task_uuid/progress", t.GetTranscodeProgress)
	}
}

// RegisterInnerApi 注册内部API
func (t *transcodeControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	// 内部API实现
	v1 := router.Group("v1/inner/transcode")
	{
		v1.POST("/tasks", t.CreateTranscodeTask)
		v1.GET("/tasks/:task_uuid", t.GetTranscodeTask)
		v1.GET("/tasks", t.ListTranscodeTasks)
		v1.PUT("/tasks/:task_uuid/status", t.UpdateTranscodeTaskStatus)
		v1.DELETE("/tasks/:task_uuid", t.CancelTranscodeTask)
		v1.GET("/tasks/:task_uuid/progress", t.GetTranscodeProgress)
	}
}

// RegisterDebugApi 注册调试API
func (t *transcodeControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {
	// 调试API实现
}

// RegisterOpsApi 注册运维API
func (t *transcodeControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	// 运维API实现
}

func (t *transcodeControllerImpl) CreateTranscodeTask(ctx *gin.Context) {
	var req cqe.TranscodeTaskCqe
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定请求参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("请求参数验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	taskDTO, err := t.transcodeApp.CreateTranscodeTask(ctx, &req)
	if err != nil {
		logger.Error("创建转码任务失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, taskDTO)
}

func (t *transcodeControllerImpl) GetTranscodeTask(ctx *gin.Context) {
	var req cqe.QueryTranscodeTaskReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		logger.Error("绑定URI参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := ctx.ShouldBindHeader(&req); err != nil {
		logger.Error("绑定Header参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("请求参数验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	taskDTO, err := t.transcodeApp.GetTranscodeTask(ctx, req.TaskUUID)
	if err != nil {
		logger.Error("获取转码任务失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, taskDTO)
}

func (t *transcodeControllerImpl) ListTranscodeTasks(ctx *gin.Context) {
	var req cqe.ListTranscodeTasksReq
	
	// 手动获取Header参数
	userUUID := ctx.GetHeader("X-User-UUID")
	if userUUID == "" {
		logger.Error("缺少必需的Header参数", map[string]interface{}{
			"header": "X-User-UUID",
		})
		restapi.Failed(ctx, errno.ErrUserUUIDRequired)
		return
	}
	req.UserUUID = userUUID

	// 设置默认分页参数
	req.Page = 1
	req.Size = 10

	// 绑定查询参数（不包含header字段的验证）
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logger.Error("绑定查询参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("请求参数验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	tasks, total, err := t.transcodeApp.ListTranscodeTasks(ctx, req.UserUUID, req.Page, req.Size)
	if err != nil {
		logger.Error("获取转码任务列表失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	// 构建响应数据
	totalPages := int(total) / req.Size
	if int(total)%req.Size > 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"tasks":       tasks,
		"total":       total,
		"page":        req.Page,
		"size":        req.Size,
		"total_pages": totalPages,
	}

	restapi.Success(ctx, response)
}

func (t *transcodeControllerImpl) UpdateTranscodeTaskStatus(ctx *gin.Context) {
	var req cqe.UpdateTranscodeTaskStatusReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		logger.Error("绑定URI参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := ctx.ShouldBindHeader(&req); err != nil {
		logger.Error("绑定Header参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定JSON参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("请求参数验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	err := t.transcodeApp.UpdateTranscodeTaskStatus(ctx, req.TaskUUID, req.Status, "")
	if err != nil {
		logger.Error("更新转码任务状态失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, map[string]interface{}{
		"message": "状态更新成功",
	})
}

func (t *transcodeControllerImpl) CancelTranscodeTask(ctx *gin.Context) {
	var req cqe.CancelTranscodeTaskReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		logger.Error("绑定URI参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := ctx.ShouldBindHeader(&req); err != nil {
		logger.Error("绑定Header参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("请求参数验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	err := t.transcodeApp.CancelTranscodeTask(ctx, req.TaskUUID)
	if err != nil {
		logger.Error("取消转码任务失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, map[string]interface{}{
		"message": "任务取消成功",
	})
}

func (t *transcodeControllerImpl) GetTranscodeProgress(ctx *gin.Context) {
	var req cqe.GetTranscodeProgressReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		logger.Error("绑定URI参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := ctx.ShouldBindHeader(&req); err != nil {
		logger.Error("绑定Header参数失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("请求参数验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	progress, err := t.transcodeApp.GetTranscodeProgress(ctx, req.TaskUUID)
	if err != nil {
		logger.Error("获取转码进度失败", map[string]interface{}{
			"error": err.Error(),
		})
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, map[string]interface{}{
		"task_uuid": req.TaskUUID,
		"progress":  progress,
	})
}
