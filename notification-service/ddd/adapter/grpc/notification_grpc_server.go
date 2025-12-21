package grpc

import (
	"context"
	"errors"
	"fmt"
	"os"

	notificationpb "notification-service/proto/notification"

	"notification-service/ddd/application/app"
	"notification-service/ddd/application/cqe"
	"notification-service/pkg/errno"
	"notification-service/pkg/logger"
)

// NotificationGrpcServer implements the gRPC NotificationService.
type NotificationGrpcServer struct {
	notificationpb.UnimplementedNotificationServiceServer
	app app.NotificationApp
}

// NewNotificationGrpcServer creates a new gRPC server implementation.
func NewNotificationGrpcServer(notificationApp app.NotificationApp) *NotificationGrpcServer {
	return &NotificationGrpcServer{
		app: notificationApp,
	}
}

// CreateNotification accepts a gRPC request and delegates to the application layer.
func (s *NotificationGrpcServer) CreateNotification(ctx context.Context, req *notificationpb.CreateNotificationRequest) (*notificationpb.CreateNotificationResponse, error) {
	if s.app == nil {
		logger.WithContext(ctx).Errorf("notification app not initialised for gRPC server")
		return &notificationpb.CreateNotificationResponse{
			Success: false,
			Message: "service unavailable",
		}, nil
	}

	if req == nil {
		return &notificationpb.CreateNotificationResponse{
			Success: false,
			Message: "request is nil",
		}, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	logger.WithContext(ctx).Infof(
		"CreateNotification handled by instance=%s user_uuid=%s type=%s title=%s",
		hostname, req.GetUserUuid(), req.GetType(), req.GetTitle(),
	)

	createReq := &cqe.CreateNotificationReq{
		UserUUID:  req.GetUserUuid(),
		Type:      req.GetType(),
		Title:     req.GetTitle(),
		Content:   req.GetContent(),
		ExtraJSON: req.GetExtraJson(),
	}

	if !createReq.Validate() {
		return &notificationpb.CreateNotificationResponse{
			Success: false,
			Message: errno.ErrParameterInvalid.Message,
		}, nil
	}

	if err := s.app.Create(ctx, createReq); err != nil {
		logger.WithContext(ctx).Errorf("CreateNotification failed user_uuid=%s type=%s title=%s error=%v",
			createReq.UserUUID, createReq.Type, createReq.Title, err)
		var bizErr errno.BizError
		if errors.As(err, &bizErr) {
			return &notificationpb.CreateNotificationResponse{
				Success: false,
				Message: bizErr.Message(),
			}, nil
		}
		return &notificationpb.CreateNotificationResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create notification: %v", err),
		}, nil
	}

	return &notificationpb.CreateNotificationResponse{
		Success: true,
		Message: "ok",
	}, nil
}
