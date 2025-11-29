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
			logger.Warnf("failed to connect video-service address=%s error=%s", address, err.Error())
		}
		singletonVideoServiceClient = client
	})
	return singletonVideoServiceClient
}

func (c *VideoServiceClient) connect() error {
	if c.address == "" {
		return fmt.Errorf("video-service address is empty")
	}
	logger.Infof("Connecting to video-service address=%s", c.address)
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
	logger.Infof("Connected to video-service address=%s", c.address)
	return nil
}

func (c *VideoServiceClient) reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	logger.Infof("Reconnecting to video-service address=%s", c.address)
	return c.connect()
}

func (c *VideoServiceClient) UpdateTranscodeResult(ctx context.Context, videoUUID, taskUUID, status, videoURL, errMsg string, durationSec int32, sizeBytes int64) (*videopb.UpdateTranscodeResultResponse, error) {
	if c.client == nil {
		if err := c.connect(); err != nil {
			logger.Errorf("video-service init failed address=%s video_uuid=%s task_uuid=%s error=%v", c.address, videoUUID, taskUUID, err)
			return nil, fmt.Errorf("video service unavailable: %w", err)
		}
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
	logger.Infof("calling video-service UpdateTranscodeResult address=%s status=%s video_uuid=%s task_uuid=%s url=%s", c.address, status, videoUUID, taskUUID, videoURL)
	resp, err := c.client.UpdateTranscodeResult(grpcCtx, req)
	if err != nil {
		logger.Warnf("video-service UpdateTranscodeResult error address=%s video_uuid=%s task_uuid=%s error=%v", c.address, videoUUID, taskUUID, err)
		if c.reconnect() == nil {
			logger.Infof("retrying video-service UpdateTranscodeResult address=%s video_uuid=%s task_uuid=%s", c.address, videoUUID, taskUUID)
			resp, err = c.client.UpdateTranscodeResult(grpcCtx, req)
		}
	}
	if err == nil && resp != nil {
		logger.Infof("video-service UpdateTranscodeResult done success=%v message=%s video_uuid=%s task_uuid=%s", resp.GetSuccess(), resp.GetMessage(), videoUUID, taskUUID)
	}
	return resp, err
}

func (c *VideoServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
