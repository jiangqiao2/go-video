package grpc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "go-vedio-1/proto/user"
	"upload-service/pkg/config"
	"upload-service/pkg/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	userServiceClientOnce      sync.Once
	singletonUserServiceClient *UserServiceClient
)

// UserServiceClient gRPC客户端
type UserServiceClient struct {
	client    pb.UserServiceClient
	conn      *grpc.ClientConn
	discovery *registry.ServiceDiscovery
	timeout   time.Duration
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Timeout        time.Duration `yaml:"timeout"`
	MaxRecvMsgSize int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize int           `yaml:"max_send_msg_size"`
	RetryTimes     int           `yaml:"retry_times"`
}

// DefaultUserServiceClient 获取默认的UserServiceClient单例
func DefaultUserServiceClient() *UserServiceClient {
	userServiceClientOnce.Do(func() {
		// 获取全局配置
		cfg := config.GetGlobalConfig()

		// 创建服务发现客户端
		registryConfig := registry.RegistryConfig{
			Endpoints:      cfg.Etcd.Endpoints,
			DialTimeout:    cfg.Etcd.DialTimeout,
			RequestTimeout: cfg.Etcd.RequestTimeout,
			Username:       cfg.Etcd.Username,
			Password:       cfg.Etcd.Password,
		}

		serviceDiscovery, err := registry.NewServiceDiscovery(registryConfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to create service discovery: %v", err))
		}

		// 启动服务发现监听
		serviceDiscovery.WatchService("user-service")

		singletonUserServiceClient = &UserServiceClient{
			discovery: serviceDiscovery,
			timeout:   cfg.GRPC.Timeout,
		}

		// 初始连接
		if err := singletonUserServiceClient.connect(); err != nil {
			panic(fmt.Sprintf("Failed to connect to user service: %v", err))
		}
	})
	return singletonUserServiceClient
}

// NewUserServiceClient 创建gRPC客户端（保留向后兼容性）
func NewUserServiceClient(discovery *registry.ServiceDiscovery, config ClientConfig) (*UserServiceClient, error) {
	client := &UserServiceClient{
		discovery: discovery,
		timeout:   config.Timeout,
	}

	// 初始连接
	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	return client, nil
}

// connect 连接到user-service
func (c *UserServiceClient) connect() error {
	// 从服务发现获取服务地址
	serviceAddr, err := c.discovery.GetServiceAddress("user-service")
	if err != nil {
		return fmt.Errorf("failed to discover user-service: %w", err)
	}

	log.Printf("Connecting to user-service at: %s", serviceAddr)

	// 建立gRPC连接
	conn, err := grpc.Dial(serviceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
	)
	if err != nil {
		return fmt.Errorf("failed to dial user-service: %w", err)
	}

	c.conn = conn
	c.client = pb.NewUserServiceClient(conn)

	log.Printf("Successfully connected to user-service")
	return nil
}

// GetUserByUUID 根据UUID获取用户信息
func (c *UserServiceClient) GetUserByUUID(ctx context.Context, userUUID string) (*pb.UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &pb.GetUserByUUIDRequest{
		UserUuid: userUUID,
	}

	resp, err := c.client.GetUserByUUID(ctx, req)
	if err != nil {
		// 尝试重新连接
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

	req := &pb.ValidateUserRequest{
		UserUuid: userUUID,
	}

	resp, err := c.client.ValidateUser(ctx, req)
	if err != nil {
		// 尝试重新连接
		if c.reconnect() == nil {
			resp, err = c.client.ValidateUser(ctx, req)
		}
		if err != nil {
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

	req := &pb.GetUsersByUUIDsRequest{
		UserUuids: userUUIDs,
	}

	resp, err := c.client.GetUsersByUUIDs(ctx, req)
	if err != nil {
		// 尝试重新连接
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
	log.Println("Attempting to reconnect to user-service...")

	// 关闭旧连接
	if c.conn != nil {
		c.conn.Close()
	}

	// 重新连接
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
