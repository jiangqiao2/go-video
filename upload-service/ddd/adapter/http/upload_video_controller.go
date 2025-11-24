package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"sync"
	"upload-service/ddd/application/app"
	uploadCqe "upload-service/ddd/application/cqe"
	"upload-service/pkg/assert"
	"upload-service/pkg/errno"
	"upload-service/pkg/manager"
	"upload-service/pkg/restapi"
)

var (
	uploadVideoControllerOnce      sync.Once
	singletonUploadVideoController UploadVideoController
)

type UploadVideoControllerPlugin struct {
}

func (p *UploadVideoControllerPlugin) Name() string {
	return "uploadVideoControllerPlugin"
}
func (p *UploadVideoControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	uploadVideoControllerOnce.Do(func() {
		singletonUploadVideoController = &uploadVideoControllerImpl{
			uploadVideoApp: app.DefaultUploadVideoApp(),
		}
	})
	assert.NotNil(singletonUploadVideoController)
	return singletonUploadVideoController
}

type UploadVideoController interface {
	manager.Controller
	Init(ctx *gin.Context)
	UploadVideoChunk(ctx *gin.Context)
	MergeChunks(ctx *gin.Context)
	TestAuth(ctx *gin.Context)
	GetStoragePath(ctx *gin.Context)
	GetUploadStatus(ctx *gin.Context)
	PresignImage(ctx *gin.Context)
	UploadImage(ctx *gin.Context)
}

type uploadVideoControllerImpl struct {
	manager.Controller
	uploadVideoApp app.UploadVideoApp
}

// RegisterOpenApi 注册开放API
func (c *uploadVideoControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	v1 := router.Group("upload/v1/open")
	{
		v1.POST("/image/presign", c.PresignImage)
	}
}

// RegisterInnerApi 注册内部API
func (c *uploadVideoControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	// 内部API实现
	v1 := router.Group("upload/v1/inner")
	{
		v1.POST("/init", c.Init)
		v1.POST("/chunk", c.UploadVideoChunk)
		v1.POST("/merge", c.MergeChunks)
		v1.GET("/chunk", c.GetStoragePath)
		v1.GET("/status", c.GetUploadStatus)
		v1.GET("/test-auth", c.TestAuth)
		v1.POST("/image", c.UploadImage)
	}

}

// RegisterDebugApi 注册调试API
func (c *uploadVideoControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {
	// 调试API实现
}

// RegisterOpsApi 注册运维API
func (c *uploadVideoControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	// 运维API实现
}

// extractUserInfo 从请求头中提取用户信息
func (c *uploadVideoControllerImpl) extractUserInfo(ctx *gin.Context) (string, error) {
	userUUID := ctx.GetHeader("X-User-UUID")

	if userUUID == "" {
		return "", errno.ErrUnauthorized
	}

	return userUUID, nil
}

func (c *uploadVideoControllerImpl) Init(ctx *gin.Context) {
	// 提取用户信息
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	var cqe uploadCqe.UploadVideoInitReq
	if err := ctx.ShouldBindJSON(&cqe); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// 将用户信息注入到请求中
	cqe.UserUUID = userUUID
	result, err := c.uploadVideoApp.UploadVideoInit(context.Background(), &cqe)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

func (c *uploadVideoControllerImpl) TestAuth(ctx *gin.Context) {
	// 提取用户信息
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// 返回测试信息
	response := map[string]interface{}{
		"message":   "JWT鉴权测试成功",
		"user_uuid": userUUID,
		"timestamp": "2025-09-14T14:42:10+08:00",
		"service":   "upload-service",
	}

	restapi.Success(ctx, response)
}

func (c *uploadVideoControllerImpl) UploadVideoChunk(ctx *gin.Context) {
	// 提取用户信息
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	var cqe uploadCqe.UploadChunkReq
	if err := ctx.ShouldBindJSON(&cqe); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// 将用户信息注入到请求中
	cqe.UserUUID = userUUID

	result, err := c.uploadVideoApp.UploadVideoChunk(context.Background(), &cqe)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

func (c *uploadVideoControllerImpl) MergeChunks(ctx *gin.Context) {
	// 提取用户信息
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	var cqe uploadCqe.MergeChunkReq
	if err := ctx.ShouldBindJSON(&cqe); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// 将用户信息注入到请求中
	cqe.UserUUID = userUUID

	result, err := c.uploadVideoApp.MergeChunks(context.Background(), &cqe)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

func (c *uploadVideoControllerImpl) GetStoragePath(ctx *gin.Context) {
	var req uploadCqe.UploadVideoStoragePathReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.uploadVideoApp.QueryStoragePath(ctx, &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *uploadVideoControllerImpl) GetUploadStatus(ctx *gin.Context) {
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req uploadCqe.UploadVideoStatusReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	req.UserUUID = userUUID
	res, err := c.uploadVideoApp.QueryUploadStatus(ctx, &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *uploadVideoControllerImpl) PresignImage(ctx *gin.Context) {
	var req uploadCqe.PresignImageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}
	req.Normalize()
	if uuid, err := c.extractUserInfo(ctx); err == nil {
		req.UserUUID = uuid
	}
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.uploadVideoApp.PresignImage(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *uploadVideoControllerImpl) UploadImage(ctx *gin.Context) {
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "file"))
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		restapi.Failed(ctx, errno.NewBizError(errno.ErrInternalServer, err))
		return
	}
	defer file.Close()
	category := ctx.PostForm("category")
	contentType := fileHeader.Header.Get("Content-Type")
	res, err := c.uploadVideoApp.UploadImage(context.Background(), userUUID, fileHeader.Filename, category, contentType, file, fileHeader.Size)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}
