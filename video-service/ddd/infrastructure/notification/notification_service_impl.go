package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	notificationpb "notification-service/proto/notification"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"video-service/ddd/domain/gateway"
	"video-service/pkg/config"
	"video-service/pkg/grpcutil"
	"video-service/pkg/logger"
)

type notificationServiceImpl struct {
	client  notificationpb.NotificationServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	address string
}

type noopNotificationService struct{}

func (noopNotificationService) NotifyVideoPublished(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

var (
	notificationOnce sync.Once
	singleton        gateway.NotificationService
)

// DefaultNotificationService 返回基于配置的通知 gRPC 客户端实现。
func DefaultNotificationService() gateway.NotificationService {
	notificationOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			logger.Warn("notification service: global config not initialised, notifications disabled")
			singleton = noopNotificationService{}
			return
		}
		dep := cfg.Dependencies.NotificationService
		address := resolveAddress(
			dep.Address,
			dep.Host,
			dep.Port,
			dep.ServiceName,
			dep.Port,
		)
		timeout := dep.Timeout
		if timeout <= 0 {
			timeout = cfg.GRPC.Timeout
		}
		if timeout <= 0 {
			timeout = 3 * time.Second
		}
		client := &notificationServiceImpl{
			address: address,
			timeout: timeout,
		}
		if err := client.connect(); err != nil {
			logger.Warnf("failed to connect notification-service, will retry later error=%s", err.Error())
		}
		singleton = client
	})
	if singleton == nil {
		singleton = noopNotificationService{}
	}
	return singleton
}

// resolveAddress 根据 address/host/服务名 构造 gRPC 地址（host:port）。
func resolveAddress(addr, host string, port int, serviceName string, defaultPort int) string {
	if addr != "" {
		return addr
	}
	if host != "" {
		if defaultPort > 0 && port <= 0 {
			port = defaultPort
		}
		return fmt.Sprintf("%s:%d", host, port)
	}
	if serviceName == "" {
		if defaultPort <= 0 {
			return ""
		}
		return fmt.Sprintf("localhost:%d", defaultPort)
	}
	if defaultPort > 0 && port <= 0 {
		port = defaultPort
	}
	return fmt.Sprintf("%s:%d", serviceName, port)
}

func (s *notificationServiceImpl) connect() error {
	if s.address == "" {
		return fmt.Errorf("notification-service address is empty")
	}
	conn, err := grpc.Dial(
		s.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(s.timeout),
		grpc.WithChainUnaryInterceptor(grpcutil.UnaryClientRequestIDInterceptor),
	)
	if err != nil {
		return fmt.Errorf("dial notification-service: %w", err)
	}
	s.conn = conn
	s.client = notificationpb.NewNotificationServiceClient(conn)
	return nil
}

// NotifyVideoPublished 在视频发布成功后通过 gRPC 调用通知服务创建一条通知。
func (s *notificationServiceImpl) NotifyVideoPublished(ctx context.Context, userUUID, videoUUID, title string) error {
	if userUUID == "" || videoUUID == "" {
		return fmt.Errorf("user_uuid or video_uuid is empty")
	}
	if s.client == nil {
		if err := s.connect(); err != nil {
			return fmt.Errorf("notification-service unavailable: %w", err)
		}
	}

	content := fmt.Sprintf("你的视频《%s》已经发布成功，可以去个人中心查看", title)
	extra := map[string]string{
		"video_uuid": videoUUID,
	}
	extraJSON, err := json.Marshal(extra)
	if err != nil {
		return err
	}

	req := &notificationpb.CreateNotificationRequest{
		UserUuid:  userUUID,
		Type:      "video_published",
		Title:     "视频发布成功",
		Content:   content,
		ExtraJson: string(extraJSON),
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.CreateNotification(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil || !resp.Success {
		msg := ""
		if resp != nil {
			msg = resp.Message
		}
		return fmt.Errorf("notification-service create notification failed: %s", msg)
	}
	return nil
}
