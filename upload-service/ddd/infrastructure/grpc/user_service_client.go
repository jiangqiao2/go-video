package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"upload-service/pkg/config"
	"upload-service/pkg/grpcutil"
	"upload-service/pkg/logger"
	"upload-service/pkg/observability"
	pb "user-service/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultUserServicePort = 9091

var (
	userServiceClientOnce      sync.Once
	singletonUserServiceClient *UserServiceClient
)

// UserServiceClient gRPC客户端，支持直连或 k8s DNS 服务名
type UserServiceClient struct {
	client  pb.UserServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	address string
}

// ClientConfig gRPC客户端配置
type ClientConfig struct {
	Timeout        time.Duration `yaml:"timeout"`
	MaxRecvMsgSize int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize int           `yaml:"max_send_msg_size"`
	RetryTimes     int           `yaml:"retry_times"`
}

// DefaultUserServiceClient 获取默认的UserServiceClient单例
func DefaultUserServiceClient() *UserServiceClient {
	userServiceClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			logger.Fatal("global config is not initialised")
			return
		}

		address := resolveUserAddress(
			cfg.Dependencies.UserService.Address,
			cfg.Dependencies.UserService.Host,
			cfg.Dependencies.UserService.Port,
			cfg.Dependencies.UserService.ServiceName,
			defaultUserServicePort,
		)

		timeout := cfg.GRPC.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}

		client := &UserServiceClient{
			timeout: timeout,
			address: address,
		}

		if err := client.connect(); err != nil {
			logger.Fatal(fmt.Sprintf("failed to connect to user-service: %v", err))
			return
		}

		singletonUserServiceClient = client
	})
	return singletonUserServiceClient
}

// NewUserServiceClient 创建gRPC客户端
func NewUserServiceClient(cfg ClientConfig) (*UserServiceClient, error) {
	globalCfg := config.GetGlobalConfig()

	address := resolveUserAddress(
		"",
		"",
		0,
		"user-service",
		defaultUserServicePort,
	)
	if globalCfg != nil {
		address = resolveUserAddress(
			globalCfg.Dependencies.UserService.Address,
			globalCfg.Dependencies.UserService.Host,
			globalCfg.Dependencies.UserService.Port,
			globalCfg.Dependencies.UserService.ServiceName,
			defaultUserServicePort,
		)
	}

	timeout := cfg.Timeout
	if timeout <= 0 && globalCfg != nil {
		timeout = globalCfg.GRPC.Timeout
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client := &UserServiceClient{
		timeout: timeout,
		address: address,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to user-service: %w", err)
	}

	return client, nil
}

// connect 建立到user-service的连接
func (c *UserServiceClient) connect() error {
	if c.address == "" {
		return fmt.Errorf("user-service address is empty")
	}

	logger.Infof("Connecting to user-service address=%s", c.address)

	conn, err := grpc.Dial(c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
		grpc.WithChainUnaryInterceptor(
			grpcutil.UnaryClientRequestIDInterceptor,
			observability.GRPCClientTracingInterceptor("upload-service"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to dial user-service: %w", err)
	}

	c.conn = conn
	c.client = pb.NewUserServiceClient(conn)

	logger.Infof("Connected to user-service address=%s", c.address)
	return nil
}

// GetUserByUUID 根据UUID获取用户信息
func (c *UserServiceClient) GetUserByUUID(ctx context.Context, userUUID string) (*pb.UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &pb.GetUserByUUIDRequest{UserUuid: userUUID}

	if c.client == nil {
		if err := c.connect(); err != nil {
			return nil, fmt.Errorf("user-service unavailable: %w", err)
		}
	}
	resp, err := c.client.GetUserByUUID(ctx, req)
	if err != nil {
		if c.reconnect() == nil {
			resp, err = c.client.GetUserByUUID(ctx, req)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get user by UUID: %w", err)
		}
	}

	if !resp.Success {
		return nil, fmt.Errorf("user service error: %s", resp.Message)
	}

	return resp.User, nil
}

// ValidateUser 验证用户是否存在
func (c *UserServiceClient) ValidateUser(ctx context.Context, userUUID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &pb.ValidateUserRequest{UserUuid: userUUID}

	if c.client == nil {
		if err := c.connect(); err != nil {
			logger.Errorf("user-service unavailable: %v", err)
			return false, fmt.Errorf("user-service unavailable: %w", err)
		}
	}
	resp, err := c.client.ValidateUser(ctx, req)
	if err != nil {
		if c.reconnect() == nil {
			resp, err = c.client.ValidateUser(ctx, req)
		}
		if err != nil {
			logger.Errorf("failed to validate user: %w", err)
			return false, fmt.Errorf("failed to validate user: %w", err)
		}
	}

	if !resp.Success {
		return false, fmt.Errorf("user service error: %s", resp.Message)
	}

	return resp.Exists, nil
}

// GetUsersByUUIDs 批量获取用户信息
func (c *UserServiceClient) GetUsersByUUIDs(ctx context.Context, userUUIDs []string) ([]*pb.UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &pb.GetUsersByUUIDsRequest{UserUuids: userUUIDs}

	if c.client == nil {
		if err := c.connect(); err != nil {
			return nil, fmt.Errorf("user-service unavailable: %w", err)
		}
	}
	resp, err := c.client.GetUsersByUUIDs(ctx, req)
	if err != nil {
		if c.reconnect() == nil {
			resp, err = c.client.GetUsersByUUIDs(ctx, req)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get users by UUIDs: %w", err)
		}
	}

	if !resp.Success {
		return nil, fmt.Errorf("user service error: %s", resp.Message)
	}

	return resp.Users, nil
}

// reconnect 重新连接
func (c *UserServiceClient) reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}

	logger.Infof("Reconnecting to user-service...")
	return c.connect()
}

// Close 关闭客户端连接
func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected 检查连接状态
func (c *UserServiceClient) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.GetState().String() == "READY"
}

func resolveUserAddress(addr, host string, port int, serviceName string, defaultPort int) string {
	if addr != "" {
		return addr
	}
	if host != "" {
		if port <= 0 {
			port = defaultPort
		}
		return fmt.Sprintf("%s:%d", host, port)
	}
	if serviceName == "" {
		return fmt.Sprintf("localhost:%d", defaultPort)
	}
	if port <= 0 {
		port = defaultPort
	}
	return fmt.Sprintf("%s:%d", serviceName, port)
}
