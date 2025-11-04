package grpc

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

var (
	uploadServiceClientOnce      sync.Once
	singletonUploadServiceClient *UploadServiceClient
)

// UploadServiceClient gRPC客户端
type UploadServiceClient struct {
	client    uploadpb.UploadServiceClient
	conn      *grpc.ClientConn
	discovery *registry.ServiceDiscovery
	timeout   time.Duration

// ClientConfig 客户端配置
type ClientConfig struct {
	Timeout        time.Duration `yaml:"timeout"`
	MaxRecvMsgSize int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize int           `yaml:"max_send_msg_size"`
	RetryTimes     int           `yaml:"retry_times"`
}

// DefaultUploadServiceClient 获取默认的gRPC客户端（单例模式）
func DefaultUploadServiceClient() *UploadServiceClient {
	uploadServiceClientOnce.Do(func() {
	uploadServiceClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			logger.Fatal("全局配置未初始化")
			return
		}
		// 创建服务发现
		registryConfig := registry.RegistryConfig{
			Endpoints:      cfg.Etcd.Endpoints,
			DialTimeout:    cfg.Etcd.DialTimeout,
			RequestTimeout: cfg.Etcd.RequestTimeout,
			Username:       cfg.Etcd.Username,
			Password:       cfg.Etcd.Password,
		}

		discovery, err := registry.NewServiceDiscovery(registryConfig)
		discovery, err := registry.NewServiceDiscovery(registryConfig)
		if err != nil {
			logger.Fatal(fmt.Sprintf("创建服务发现失败: %v", err))
			return
		}

		timeout := cfg.GRPCClient.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}

		client := &UploadServiceClient{
			discovery: discovery,
			timeout:   timeout,
		client := &UploadServiceClient{
			discovery: discovery,
			timeout:   timeout,
		}

		// 初始连接
		err = client.connect()
		if err != nil {
			logger.Fatal(fmt.Sprintf("连接upload-service失败: %v", err))
			return
		}
			return
		singletonUploadServiceClient = client
	})
		singletonUploadServiceClient = client
	return singletonUploadServiceClient

	return singletonUploadServiceClient
}

	client := &UploadServiceClient{
		discovery: discovery,
		timeout:   config.Timeout,
	}

	// 初始连接
	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to upload service: %w", err)
	}

	return client, nil
	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to upload service: %w", err)
	}
	// 从服务发现获取服务地址
	serviceAddr, err := c.discovery.GetServiceAddress("upload-service-grpc")
	if err != nil {
		return fmt.Errorf("failed to discover upload-service-grpc: %w", err)
	}

	logger.Info("正在连接upload-service", map[string]interface{}{
		"address": serviceAddr,
	})

	// 建立gRPC连接
	conn, err := grpc.Dial(serviceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
	)
	if err != nil {
		return fmt.Errorf("failed to dial upload-service: %w", err)
	}

	c.conn = conn
	c.client = uploadpb.NewUploadServiceClient(conn)

	logger.Info("成功连接到upload-service", map[string]interface{}{
		"address": serviceAddr,
	})
	return nil
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
		logger.Error("调用UpdateTranscodeStatus失败", map[string]interface{}{
			"video_uuid": videoUUID,
	req := &uploadpb.UpdateTranscodeStatusRequest{
		VideoUuid:          videoUUID,
		TranscodeTaskUuid:  transcodeTaskUUID,
		Status:             status,
		VideoUrl:           videoURL,
		HlsPlaylistUrl:     hlsPlaylistURL,
		ErrorMessage:       errorMessage,
	}
		"status":     status,
	// 创建带超时的上下文
	grpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return resp, nil
}

// Close 关闭gRPC连接
func (c *UploadServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
