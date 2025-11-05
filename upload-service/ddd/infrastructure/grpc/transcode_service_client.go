package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	transcodepb "go-vedio-1/proto/transcode"

	"upload-service/pkg/config"
	"upload-service/pkg/logger"
	"upload-service/pkg/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	transcodeClientOnce      sync.Once
	singletonTranscodeClient *TranscodeServiceClient
)

// TranscodeServiceClient wraps gRPC interactions with the transcode service.
type TranscodeServiceClient struct {
	client    transcodepb.TranscodeServiceClient
	conn      *grpc.ClientConn
	discovery *registry.ServiceDiscovery
	timeout   time.Duration
}

// DefaultTranscodeServiceClient returns a singleton configured via global config.
func DefaultTranscodeServiceClient() *TranscodeServiceClient {
    transcodeClientOnce.Do(func() {
        cfg := config.GetGlobalConfig()
        registryConfig := registry.RegistryConfig{
            Endpoints:      cfg.Etcd.Endpoints,
            DialTimeout:    cfg.Etcd.DialTimeout,
            RequestTimeout: cfg.Etcd.RequestTimeout,
            Username:       cfg.Etcd.Username,
            Password:       cfg.Etcd.Password,
        }

        serviceDiscovery, err := registry.NewServiceDiscovery(registryConfig)
        if err != nil {
            logger.Warn("创建服务发现失败，转码客户端将延迟连接", map[string]interface{}{"error": err.Error()})
            serviceDiscovery = nil
        }

        if serviceDiscovery != nil {
            serviceDiscovery.WatchService("transcode-service")
        }

		client := &TranscodeServiceClient{
			discovery: serviceDiscovery,
			timeout:   cfg.GRPC.Timeout,
		}

        // 尝试连接，但如果失败不阻塞服务启动
        if err := client.connect(); err != nil {
            logger.Warn("连接转码服务失败，稍后将重试", map[string]interface{}{"error": err.Error()})
        }

		singletonTranscodeClient = client
	})
	return singletonTranscodeClient
}

// NewTranscodeServiceClient creates a client using provided discovery and config.
func NewTranscodeServiceClient(discovery *registry.ServiceDiscovery, cfg ClientConfig) (*TranscodeServiceClient, error) {
	client := &TranscodeServiceClient{
		discovery: discovery,
		timeout:   cfg.Timeout,
	}
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to transcode-service: %w", err)
	}
	return client, nil
}

func (c *TranscodeServiceClient) connect() error {
	serviceAddr, err := c.discovery.GetServiceAddress("transcode-service")
	if err != nil {
		return fmt.Errorf("discover transcode-service: %w", err)
	}

	conn, err := grpc.Dial(
		serviceAddr,
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
	logger.Info("Reconnecting to transcode-service...", nil)
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

    // 如果尚未建立连接，先尝试建立
    if c.client == nil {
        if c.discovery == nil {
            return nil, fmt.Errorf("service discovery unavailable for transcode-service")
        }
        if err := c.connect(); err != nil {
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

    // 如果尚未建立连接，先尝试建立
    if c.client == nil {
        if c.discovery == nil {
            return nil, fmt.Errorf("service discovery unavailable for transcode-service")
        }
        if err := c.connect(); err != nil {
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
