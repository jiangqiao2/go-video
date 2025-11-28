package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	transcodepb "transcode-service/proto/transcode"

	"upload-service/pkg/config"
	"upload-service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	transcodeClientOnce      sync.Once
	singletonTranscodeClient *TranscodeServiceClient
)

// TranscodeServiceClient wraps gRPC interactions with the transcode service.
type TranscodeServiceClient struct {
	client  transcodepb.TranscodeServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	address string
}

// DefaultTranscodeServiceClient returns a singleton configured via global config.
func DefaultTranscodeServiceClient() *TranscodeServiceClient {
	transcodeClientOnce.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			logger.Fatal("global config is not initialised")
			return
		}

		address := resolveTranscodeAddress(
			cfg.Dependencies.TranscodeService.Address,
			cfg.Dependencies.TranscodeService.Host,
			cfg.Dependencies.TranscodeService.Port,
			cfg.Dependencies.TranscodeService.ServiceName,
			cfg.Dependencies.TranscodeService.Port,
		)

		timeout := cfg.GRPC.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}

		client := &TranscodeServiceClient{
			timeout: timeout,
			address: address,
		}

		if err := client.connect(); err != nil {
			logger.Warnf("failed to connect transcode-service, will retry later error=%s", err.Error())
		}

		singletonTranscodeClient = client
	})
	return singletonTranscodeClient
}

// NewTranscodeServiceClient creates a client using provided config.
func NewTranscodeServiceClient(cfg ClientConfig) (*TranscodeServiceClient, error) {
	globalCfg := config.GetGlobalConfig()

	address := resolveTranscodeAddress("", "", 0, "transcode-service", 0)
	if globalCfg != nil {
		address = resolveTranscodeAddress(
			globalCfg.Dependencies.TranscodeService.Address,
			globalCfg.Dependencies.TranscodeService.Host,
			globalCfg.Dependencies.TranscodeService.Port,
			globalCfg.Dependencies.TranscodeService.ServiceName,
			globalCfg.Dependencies.TranscodeService.Port,
		)
	}

	timeout := cfg.Timeout
	if timeout <= 0 && globalCfg != nil {
		timeout = globalCfg.GRPC.Timeout
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client := &TranscodeServiceClient{
		timeout: timeout,
		address: address,
	}
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to transcode-service: %w", err)
	}
	return client, nil
}

func (c *TranscodeServiceClient) connect() error {
	if c.address == "" {
		return fmt.Errorf("transcode-service address is empty")
	}

	conn, err := grpc.Dial(
		c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(c.timeout),
	)
	if err != nil {
		return fmt.Errorf("dial transcode-service: %w", err)
	}

	c.conn = conn
	c.client = transcodepb.NewTranscodeServiceClient(conn)
	return nil
}

func (c *TranscodeServiceClient) reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	logger.Infof("Reconnecting to transcode-service...")
	return c.connect()
}

// Close closes the underlying gRPC connection.
func (c *TranscodeServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateTranscodeTask enqueues a new transcode job.
func (c *TranscodeServiceClient) CreateTranscodeTask(ctx context.Context, req *transcodepb.CreateTranscodeTaskRequest) (*transcodepb.CreateTranscodeTaskResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.client == nil {
		if err := c.connect(); err != nil {
			logger.Errorf("CreateTranscodeTask %v error:%v", req.VideoUuid, err)
			return nil, fmt.Errorf("transcode-service unavailable: %w", err)
		}
	}

	resp, err := c.client.CreateTranscodeTask(ctx, req)
	if err != nil {
		if c.reconnect() == nil {
			resp, err = c.client.CreateTranscodeTask(ctx, req)
		}
	}
	return resp, err
}

// GetTranscodeTask fetches task status by uuid.
func (c *TranscodeServiceClient) GetTranscodeTask(ctx context.Context, req *transcodepb.GetTranscodeTaskRequest) (*transcodepb.GetTranscodeTaskResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.client == nil {
		if err := c.connect(); err != nil {
			logger.Errorf("GetTranscodeTask %v error:%v", req.GetTaskUuid(), err)
			return nil, fmt.Errorf("transcode-service unavailable: %w", err)
		}
	}

	resp, err := c.client.GetTranscodeTask(ctx, req)
	if err != nil {
		if c.reconnect() == nil {
			resp, err = c.client.GetTranscodeTask(ctx, req)
		}
	}
	return resp, err
}

func resolveTranscodeAddress(addr, host string, port int, serviceName string, defaultPort int) string {
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
