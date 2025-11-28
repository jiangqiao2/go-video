package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	videopb "video-service/proto/video"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	videoServiceClientOnce      sync.Once
	singletonVideoServiceClient *VideoServiceClient
)

type VideoServiceClient struct {
	client  videopb.VideoServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	address string
}

func DefaultVideoServiceClient() *VideoServiceClient {
	videoServiceClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			logger.Fatal("global config is not initialised")
			return
		}
		address := resolveAddress(
			cfg.Dependencies.VideoService.Address,
			cfg.Dependencies.VideoService.Host,
			cfg.Dependencies.VideoService.Port,
			cfg.Dependencies.VideoService.ServiceName,
			cfg.Dependencies.VideoService.Port,
		)
		timeout := cfg.Dependencies.VideoService.Timeout
		if timeout <= 0 {
			timeout = cfg.GRPCClient.Timeout
		}
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		client := &VideoServiceClient{timeout: timeout, address: address}
		if err := client.connect(); err != nil {
			logger.Warn("failed to connect video-service, will retry later", map[string]interface{}{"error": err.Error()})
		}
		singletonVideoServiceClient = client
	})
	return singletonVideoServiceClient
}

func (c *VideoServiceClient) connect() error {
	if c.address == "" {
		return fmt.Errorf("video-service address is empty")
	}
	conn, err := grpc.Dial(
		c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
	)
	if err != nil {
		return fmt.Errorf("dial video-service: %w", err)
	}
	c.conn = conn
	c.client = videopb.NewVideoServiceClient(conn)
	return nil
}

func (c *VideoServiceClient) UpdateTranscodeResult(ctx context.Context, videoUUID, taskUUID, status, videoURL, errMsg string, durationSec int32, sizeBytes int64) (*videopb.UpdateTranscodeResultResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("video service client not initialised")
	}
	req := &videopb.UpdateTranscodeResultRequest{
		VideoUuid:   videoUUID,
		TaskUuid:    taskUUID,
		Status:      status,
		VideoUrl:    videoURL,
		ErrorMsg:    errMsg,
		DurationSec: durationSec,
		SizeBytes:   sizeBytes,
	}
	grpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.UpdateTranscodeResult(grpcCtx, req)
}

func (c *VideoServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
