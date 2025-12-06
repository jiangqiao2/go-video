package http

import (
	"context"
	"time"

	grpcClient "upload-service/ddd/infrastructure/grpc"
	"upload-service/pkg/errno"
	"upload-service/pkg/manager"
	"upload-service/pkg/restapi"

	"github.com/gin-gonic/gin"
)

// TestGRPCController 测试gRPC通信的控制器
type TestGRPCController struct {
	userServiceClient *grpcClient.UserServiceClient
}

// TestGRPCControllerPlugin 测试gRPC控制器插件
type TestGRPCControllerPlugin struct{}

func (p *TestGRPCControllerPlugin) Name() string {
	return "testGRPCControllerPlugin"
}

func (p *TestGRPCControllerPlugin) MustCreateService(deps *manager.Dependencies) manager.Service {
	return &TestGRPCController{
		userServiceClient: deps.UserServiceClient,
	}
}

func (c *TestGRPCController) GetName() string {
	return "TestGRPCController"
}

func (c *TestGRPCController) RegisterRoutes(router *gin.Engine) {
	// 测试gRPC通信的接口
	router.GET("/api/v1/test/grpc/user/:uuid", c.TestGetUser)
	router.GET("/api/v1/test/grpc/validate/:uuid", c.TestValidateUser)
}

// TestGetUser 测试获取用户信息
func (c *TestGRPCController) TestGetUser(ctx *gin.Context) {
	userUUID := ctx.Param("uuid")
	if userUUID == "" {
		restapi.Failed(ctx, errno.ErrParameterInvalid)
		return
	}

	// 创建上下文
	grpcCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	// 调用gRPC服务
	userInfo, err := c.userServiceClient.GetUserByUUID(grpcCtx, userUUID)
	if err != nil {
		restapi.Failed(ctx, errno.NewBizError(errno.ErrInternalServer, err))
		return
	}

	restapi.Success(ctx, gin.H{
		"message":   "gRPC调用成功",
		"user_info": userInfo,
		"method":    "GetUserByUUID",
		"uuid":      userUUID,
	})
}

// TestValidateUser 测试验证用户
func (c *TestGRPCController) TestValidateUser(ctx *gin.Context) {
	userUUID := ctx.Param("uuid")
	if userUUID == "" {
		restapi.Failed(ctx, errno.ErrParameterInvalid)
		return
	}

	// 创建上下文
	grpcCtx, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	// 调用gRPC服务
	exists, err := c.userServiceClient.ValidateUser(grpcCtx, userUUID)
	if err != nil {
		restapi.Failed(ctx, errno.NewBizError(errno.ErrInternalServer, err))
		return
	}

	restapi.Success(ctx, gin.H{
		"message": "gRPC调用成功",
		"exists":  exists,
		"method":  "ValidateUser",
		"uuid":    userUUID,
	})
}

func init() {
	manager.RegisterServicePlugin(&TestGRPCControllerPlugin{})
}
