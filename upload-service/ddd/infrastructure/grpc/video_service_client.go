package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	videopb "video-service/proto/video"

	"upload-service/pkg/config"
	"upload-service/pkg/grpcutil"
	"upload-service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	videoClientOnce      sync.Once
	singletonVideoClient *VideoServiceClient
)

type VideoServiceClient struct {
	conn    *grpc.ClientConn
	client  videopb.VideoServiceClient
	timeout time.Duration
	address string
}

func DefaultVideoServiceClient() *VideoServiceClient {
	videoClientOnce.Do(func() {
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
		timeout := cfg.GRPC.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		client := &VideoServiceClient{timeout: timeout, address: address}
		if err := client.connect(); err != nil {
			logger.Warnf("failed to connect video-service, will retry later error=%s", err.Error())
		}
		singletonVideoClient = client
	})
	return singletonVideoClient
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
		grpc.WithChainUnaryInterceptor(grpcutil.UnaryClientRequestIDInterceptor),
	)
	if err != nil {
		return fmt.Errorf("dial video-service: %w", err)
	}
	c.conn = conn
	c.client = videopb.NewVideoServiceClient(conn)
	return nil
}

func (c *VideoServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *VideoServiceClient) Precreate(ctx context.Context, req *videopb.PrecreateRequest) (*videopb.PrecreateResponse, error) {
	if c.client == nil {
		if err := c.connect(); err != nil {
			return nil, fmt.Errorf("video-service unavailable: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.Precreate(ctx, req)
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
