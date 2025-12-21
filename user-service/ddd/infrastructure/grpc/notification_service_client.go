package grpc

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	notificationpb "notification-service/proto/notification"

	"user-service/pkg/config"
	"user-service/pkg/grpcutil"
	"user-service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	notificationClientOnce      sync.Once
	singletonNotificationClient *NotificationServiceClient
)

// NotificationServiceClient wraps gRPC interactions with notification-service.
type NotificationServiceClient struct {
	client  notificationpb.NotificationServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	address string
}

// DefaultNotificationServiceClient returns a singleton client configured via
// CONFIG_PATH/config and optional NOTIFICATION_GRPC_ADDR environment variable.
// This is mainly used for debug / load-balance tests.
func DefaultNotificationServiceClient() *NotificationServiceClient {
	notificationClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()

		timeout := 30 * time.Second
		if cfg != nil && cfg.GRPC.Timeout > 0 {
			timeout = cfg.GRPC.Timeout
		}

		address := resolveNotificationAddress(os.Getenv("NOTIFICATION_GRPC_ADDR"))

		client := &NotificationServiceClient{
			timeout: timeout,
			address: address,
		}

		if err := client.connect(); err != nil {
			// 对调试接口来说，连接失败不致命，后续调用会尝试重连。
			logger.Warnf("failed to connect notification-service, will retry later error=%v", err)
		}

		singletonNotificationClient = client
	})
	return singletonNotificationClient
}

func resolveNotificationAddress(override string) string {
	if override != "" {
		return override
	}
	// 在 k8s 集群内，使用 Service 名称即可进行 gRPC 访问。
	// 默认端口为 notification-service Service 中暴露的 gRPC 端口 9095。
	return "notification-service:9095"
}

// connect 建立到 notification-service 的连接。
func (c *NotificationServiceClient) connect() error {
	if c.address == "" {
		return fmt.Errorf("notification-service address is empty")
	}

	logger.Infof("Connecting to notification-service address=%s", c.address)

	conn, err := grpc.Dial(
		c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
		grpc.WithChainUnaryInterceptor(
			grpcutil.UnaryClientRequestIDInterceptor,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to dial notification-service: %w", err)
	}

	c.conn = conn
	c.client = notificationpb.NewNotificationServiceClient(conn)

	logger.Infof("Connected to notification-service address=%s", c.address)
	return nil
}

// reconnect 在请求失败时重连一次。
func (c *NotificationServiceClient) reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	logger.Infof("Reconnecting to notification-service...")
	return c.connect()
}

// CreateNotification 发送 CreateNotification RPC 调用。
func (c *NotificationServiceClient) CreateNotification(ctx context.Context, req *notificationpb.CreateNotificationRequest) (*notificationpb.CreateNotificationResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("notification service client is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.client == nil {
		if err := c.connect(); err != nil {
			return nil, fmt.Errorf("notification-service unavailable: %w", err)
		}
	}

	resp, err := c.client.CreateNotification(ctx, req)
	if err != nil {
		if c.reconnect() == nil {
			resp, err = c.client.CreateNotification(ctx, req)
		}
	}
	return resp, err
}

// Close 关闭 gRPC 连接。
func (c *NotificationServiceClient) Close() error {
	if c == nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
