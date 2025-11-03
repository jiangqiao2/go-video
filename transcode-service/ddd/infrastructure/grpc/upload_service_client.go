package grpcclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	uploadpb "go-vedio-1/proto/upload"

	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UploadServiceClient wraps gRPC interactions with upload-service.
type UploadServiceClient struct {
	client         uploadpb.UploadServiceClient
	conn           *grpc.ClientConn
	discovery      *registry.ServiceDiscovery
	serviceName    string
	timeout        time.Duration
	retryTimes     int
	maxRecvMsgSize int
	maxSendMsgSize int
}

// ClientConfig customises timeouts and payload limits.
type ClientConfig struct {
	Timeout        time.Duration
	MaxRecvMsgSize int
	MaxSendMsgSize int
	RetryTimes     int
}

var (
	uploadClientOnce      sync.Once
	singletonUploadClient *UploadServiceClient
)

// DefaultUploadServiceClient returns a lazily initialised singleton client.
func DefaultUploadServiceClient() *UploadServiceClient {
	uploadClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			panic("global config is not initialised")
		}

		registryConfig := registry.RegistryConfig{
			Endpoints:      cfg.Etcd.Endpoints,
			DialTimeout:    cfg.Etcd.DialTimeout,
			RequestTimeout: cfg.Etcd.RequestTimeout,
			Username:       cfg.Etcd.Username,
			Password:       cfg.Etcd.Password,
		}

		discovery, err := registry.NewServiceDiscovery(registryConfig)
		if err != nil {
			panic(fmt.Sprintf("failed to create service discovery: %v", err))
		}

		serviceName := cfg.Dependencies.UploadService.ServiceName
		if serviceName == "" {
			serviceName = "upload-service"
		}
		discovery.WatchService(serviceName)

		timeout := cfg.Dependencies.UploadService.Timeout
		if timeout <= 0 {
			timeout = cfg.GRPCClient.Timeout
		}
		if timeout <= 0 {
			timeout = 30 * time.Second
		}

		client := &UploadServiceClient{
			discovery:      discovery,
			serviceName:    serviceName,
			timeout:        timeout,
			retryTimes:     cfg.GRPCClient.RetryTimes,
			maxRecvMsgSize: cfg.GRPCClient.MaxRecvMsgSize,
			maxSendMsgSize: cfg.GRPCClient.MaxSendMsgSize,
		}

		if err := client.connect(); err != nil {
			panic(fmt.Sprintf("failed to connect to %s: %v", serviceName, err))
		}

		singletonUploadClient = client
	})

	return singletonUploadClient
}

// NewUploadServiceClient builds a client using custom discovery/configuration.
func NewUploadServiceClient(discovery *registry.ServiceDiscovery, serviceName string, cfg ClientConfig) (*UploadServiceClient, error) {
	if serviceName == "" {
		serviceName = "upload-service"
	}
	client := &UploadServiceClient{
		discovery:      discovery,
		serviceName:    serviceName,
		timeout:        cfg.Timeout,
		retryTimes:     cfg.RetryTimes,
		maxRecvMsgSize: cfg.MaxRecvMsgSize,
		maxSendMsgSize: cfg.MaxSendMsgSize,
	}
	if client.timeout <= 0 {
		client.timeout = 30 * time.Second
	}
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}
	return client, nil
}

func (c *UploadServiceClient) connect() error {
	if c.discovery == nil {
		return fmt.Errorf("service discovery is not initialised")
	}

	serviceAddr, err := c.discovery.GetServiceAddress(c.serviceName)
	if err != nil {
		return fmt.Errorf("discover %s: %w", c.serviceName, err)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
	}

	callOpts := make([]grpc.CallOption, 0, 2)
	if c.maxRecvMsgSize > 0 {
		callOpts = append(callOpts, grpc.MaxCallRecvMsgSize(c.maxRecvMsgSize))
	}
	if c.maxSendMsgSize > 0 {
		callOpts = append(callOpts, grpc.MaxCallSendMsgSize(c.maxSendMsgSize))
	}
	if len(callOpts) > 0 {
		dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(callOpts...))
	}

	conn, err := grpc.Dial(serviceAddr, dialOpts...)
	if err != nil {
		return fmt.Errorf("dial upload-service at %s: %w", serviceAddr, err)
	}

	c.conn = conn
	c.client = uploadpb.NewUploadServiceClient(conn)

	logger.Info("连接upload-service成功", map[string]interface{}{
		"address": serviceAddr,
	})

	return nil
}

func (c *UploadServiceClient) reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	logger.Info("尝试重新连接upload-service...", map[string]interface{}{
		"service": c.serviceName,
	})
	return c.connect()
}

// Close shuts down the underlying gRPC connection.
func (c *UploadServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// UpdateTranscodeStatus notifies upload-service about task outcome.
func (c *UploadServiceClient) UpdateTranscodeStatus(ctx context.Context, req *uploadpb.UpdateTranscodeStatusRequest) (*uploadpb.UpdateTranscodeStatusResponse, error) {
	attempts := c.retryTimes + 1
	var lastErr error

	for i := 0; i < attempts; i++ {
		rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
		resp, err := c.client.UpdateTranscodeStatus(rpcCtx, req)
		cancel()

		if err == nil {
			return resp, nil
		}

		lastErr = err
		logger.Warn("调用upload-service.UpdateTranscodeStatus失败", map[string]interface{}{
			"attempt": i + 1,
			"error":   err.Error(),
		})

		if i == attempts-1 {
			break
		}

		if err := c.reconnect(); err != nil {
			return nil, fmt.Errorf("reconnect upload-service failed: %w", err)
		}
	}

	return nil, lastErr
}
