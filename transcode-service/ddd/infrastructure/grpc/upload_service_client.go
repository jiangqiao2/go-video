package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"transcode-service/pkg/config"
	"transcode-service/pkg/grpcutil"
	"transcode-service/pkg/logger"
	uploadpb "upload-service/proto/upload"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	uploadServiceClientOnce      sync.Once
	singletonUploadServiceClient *UploadServiceClient
)

// UploadServiceClient gRPC客户端
type UploadServiceClient struct {
	client  uploadpb.UploadServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	address string
}

// DefaultUploadServiceClient 获取默认的gRPC客户端（单例模式）
func DefaultUploadServiceClient() *UploadServiceClient {
	uploadServiceClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			logger.Fatal("global config is not initialised")
			return
		}

		address := resolveAddress(
			cfg.Dependencies.UploadService.Address,
			cfg.Dependencies.UploadService.Host,
			cfg.Dependencies.UploadService.Port,
			cfg.Dependencies.UploadService.ServiceName,
			cfg.Dependencies.UploadService.Port,
		)

		timeout := cfg.Dependencies.UploadService.Timeout
		if timeout <= 0 {
			timeout = cfg.GRPCClient.Timeout
		}
		if timeout <= 0 {
			timeout = 30 * time.Second
		}

		client := &UploadServiceClient{
			timeout: timeout,
			address: address,
		}

		if err := client.connect(); err != nil {
			logger.Fatal(fmt.Sprintf("failed to connect upload-service: %v", err))
			return
		}

		singletonUploadServiceClient = client
	})
	return singletonUploadServiceClient
}

// connect 连接到upload-service
func (c *UploadServiceClient) connect() error {
	if c.address == "" {
		return fmt.Errorf("upload-service address is empty")
	}

	logger.Infof("Connecting to upload-service address=%s", c.address)

	conn, err := grpc.Dial(c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
		grpc.WithChainUnaryInterceptor(grpcutil.UnaryClientRequestIDInterceptor),
	)
	if err != nil {
		return fmt.Errorf("failed to dial upload-service: %w", err)
	}

	c.conn = conn
	c.client = uploadpb.NewUploadServiceClient(conn)

	logger.Infof("Connected to upload-service address=%s", c.address)
	return nil
}

// UpdateTranscodeStatus 调用上传服务更新转码状态
func (c *UploadServiceClient) UpdateTranscodeStatus(ctx context.Context, videoUUID, transcodeTaskUUID, status, videoURL, errorMessage string) (*uploadpb.UpdateTranscodeStatusResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("upload service client not initialised")
	}

	req := &uploadpb.UpdateTranscodeStatusRequest{
		VideoUuid:         videoUUID,
		TranscodeTaskUuid: transcodeTaskUUID,
		Status:            status,
		VideoUrl:          videoURL,
		ErrorMessage:      errorMessage,
	}

	// 创建带超时的上下文
	grpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.UpdateTranscodeStatus(grpcCtx, req)
	if err != nil {
		logger.Errorf("UpdateTranscodeStatus failed video_uuid=%s task_uuid=%s status=%s error=%v", videoUUID, transcodeTaskUUID, status, err)
		return nil, err
	}
	return resp, nil
}

// Close 关闭gRPC连接
func (c *UploadServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

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
		return fmt.Sprintf("localhost:%d", defaultPort)
	}
	if defaultPort > 0 && port <= 0 {
		port = defaultPort
	}
	return fmt.Sprintf("%s:%d", serviceName, port)
}
