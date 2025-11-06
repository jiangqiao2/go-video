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
}

// DefaultUploadServiceClient 获取默认的gRPC客户端（单例模式）
func DefaultUploadServiceClient() *UploadServiceClient {
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
        if err != nil {
            logger.Fatal(fmt.Sprintf("创建服务发现失败: %v", err))
            return
        }

        // 读取超时配置（优先依赖中的upload_service，其次grpc_client）
        timeout := cfg.Dependencies.UploadService.Timeout
        if timeout <= 0 {
            timeout = cfg.GRPCClient.Timeout
        }
        if timeout <= 0 {
            timeout = 30 * time.Second
        }

        client := &UploadServiceClient{
            discovery: discovery,
            timeout:   timeout,
        }

        // 初始连接
        if err = client.connect(); err != nil {
            logger.Fatal(fmt.Sprintf("连接upload-service失败: %v", err))
            return
        }

        singletonUploadServiceClient = client
    })
    return singletonUploadServiceClient
}

// connect 连接到upload-service（通过etcd服务发现）
func (c *UploadServiceClient) connect() error {
    if c.discovery == nil {
        return fmt.Errorf("service discovery unavailable for upload-service")
    }

    // 使用配置中的服务名，避免硬编码导致连到错误的实例
    cfg := config.GetGlobalConfig()
    serviceName := "upload-service"
    if cfg != nil && cfg.Dependencies.UploadService.ServiceName != "" {
        serviceName = cfg.Dependencies.UploadService.ServiceName
    }

    // 从服务发现获取对应服务地址
    serviceAddr, err := c.discovery.GetServiceAddress(serviceName)
    if err != nil {
        return fmt.Errorf("failed to discover %s: %w", serviceName, err)
    }

    logger.Info("正在连接upload-service", map[string]interface{}{
        "service": serviceName,
        "address": serviceAddr,
    })

    // 建立gRPC连接
    conn, err := grpc.Dial(serviceAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithTimeout(c.timeout),
    )
    if err != nil {
        return fmt.Errorf("failed to dial %s: %w", serviceName, err)
    }

    c.conn = conn
    c.client = uploadpb.NewUploadServiceClient(conn)

    logger.Info("成功连接到upload-service", map[string]interface{}{
        "service": serviceName,
        "address": serviceAddr,
    })
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
        logger.Error("调用UpdateTranscodeStatus失败", map[string]interface{}{
            "video_uuid": videoUUID,
            "task_uuid":  transcodeTaskUUID,
            "status":     status,
            "error":      err.Error(),
        })
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