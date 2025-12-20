package http

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"notification-service/ddd/application/app"
	"notification-service/ddd/application/cqe"
	"notification-service/pkg/errno"
	"notification-service/pkg/logger"
	"notification-service/pkg/manager"
	"notification-service/pkg/restapi"
	"notification-service/pkg/sse"
)

var (
	notificationControllerOnce sync.Once
	singletonNotificationCtrl  NotificationController
)

// NotificationControllerPlugin 将通知控制器注册到共享的 manager 中。
type NotificationControllerPlugin struct{}

func (p *NotificationControllerPlugin) Name() string {
	return "notificationController"
}

func (p *NotificationControllerPlugin) MustCreateController() manager.Controller {
	notificationControllerOnce.Do(func() {
		singletonNotificationCtrl = &notificationControllerImpl{
			app: app.DefaultNotificationApp(),
		}
	})
	return singletonNotificationCtrl
}

// NotificationController 控制器接口。
type NotificationController interface {
	manager.Controller
	List(ctx *gin.Context)
	MarkRead(ctx *gin.Context)
	Create(ctx *gin.Context)
	Stream(ctx *gin.Context)
}

type notificationControllerImpl struct {
	manager.Controller
	app app.NotificationApp
}

// RegisterOpenApi 暂无开放通知接口。
func (c *notificationControllerImpl) RegisterOpenApi(group *gin.RouterGroup) {}

// RegisterInnerApi 注册内部通知接口（Kong 网关 inner 路由访问）。
func (c *notificationControllerImpl) RegisterInnerApi(group *gin.RouterGroup) {
	v1 := group.Group("notification/v1/inner")
	{
		v1.GET("/notifications", c.List)
		v1.POST("/notifications/read", c.MarkRead)
		v1.POST("/notifications", c.Create)
		v1.GET("/notifications/stream", c.Stream)
	}
}

func (c *notificationControllerImpl) RegisterDebugApi(group *gin.RouterGroup) {}
func (c *notificationControllerImpl) RegisterOpsApi(group *gin.RouterGroup)   {}

func (c *notificationControllerImpl) extractUserUUID(ctx *gin.Context) (string, error) {
	userUUID := ctx.GetHeader("X-User-UUID")
	if userUUID == "" {
		// Fallback for SSE where headers are hard to set; user_uuid can be passed via query.
		userUUID = ctx.Query("user_uuid")
	}
	if userUUID == "" {
		// 通知服务自身不做鉴权，只校验参数是否完整。
		return "", errno.ErrParameterInvalid
	}
	return userUUID, nil
}

// List 列出当前用户的通知列表以及未读数量。
func (c *notificationControllerImpl) List(ctx *gin.Context) {
	userUUID, err := c.extractUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.ListNotificationsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}
	resp, err := c.app.ListNotifications(ctx.Request.Context(), userUUID, &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, resp)
}

// MarkRead 将指定通知标记为已读。
func (c *notificationControllerImpl) MarkRead(ctx *gin.Context) {
	userUUID, err := c.extractUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.MarkReadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}
	if err := c.app.MarkRead(ctx.Request.Context(), userUUID, &req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, gin.H{"status": "ok"})
}

// Create 通过内部接口创建一条新的通知，用于其他服务调用。
func (c *notificationControllerImpl) Create(ctx *gin.Context) {
	var req cqe.CreateNotificationReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}
	if !req.Validate() {
		restapi.Failed(ctx, errno.ErrParameterInvalid)
		return
	}
	if err := c.app.Create(ctx.Request.Context(), &req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, gin.H{"status": "ok"})
}

// Stream establishes an SSE stream for the current user's notifications.
// Frontend should listen for "notification.created"/"notification.updated"
// events and trigger a notifications refresh on each event.
func (c *notificationControllerImpl) Stream(ctx *gin.Context) {
	userUUID, err := c.extractUserUUID(ctx)
	if err != nil {
		// 缺少 user_uuid 视为参数错误，而不是鉴权失败。
		restapi.FailedWithStatus(ctx, errno.ErrParameterInvalid, http.StatusBadRequest)
		return
	}

	// Prepare SSE headers.
	w := ctx.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.WithContext(ctx.Request.Context()).Errorf("notification: SSE stream does not support flushing user_uuid=%s", userUUID)
		restapi.FailedWithStatus(ctx, errno.ErrInternalServer, http.StatusInternalServerError)
		return
	}

	events, unsubscribe := sse.DefaultHub().Subscribe(userUUID)
	defer unsubscribe()

	// Initial comment to keep some proxies happy.
	if _, err := w.Write([]byte(": ok\n\n")); err == nil {
		flusher.Flush()
	}

	// Periodic heartbeat to keep long-lived connections from timing out on proxies.
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	notify := ctx.Request.Context().Done()
	for {
		select {
		case <-notify:
			return
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			flusher.Flush()
		case ev, ok := <-events:
			if !ok {
				return
			}
			data, err := json.Marshal(ev.Data)
			if err != nil {
				continue
			}
			if _, err := w.Write([]byte("event: " + ev.Type + "\n")); err != nil {
				return
			}
			if _, err := w.Write([]byte("data: ")); err != nil {
				return
			}
			if _, err := w.Write(data); err != nil {
				return
			}
			if _, err := w.Write([]byte("\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
