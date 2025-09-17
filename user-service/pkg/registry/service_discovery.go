package registry

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServiceDiscovery 服务发现客户端
type ServiceDiscovery struct {
	client   *clientv3.Client
	services map[string][]string // serviceName -> []serviceAddr
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewServiceDiscovery 创建服务发现客户端
func NewServiceDiscovery(config RegistryConfig) (*ServiceDiscovery, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout,
		Username:    config.Username,
		Password:    config.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sd := &ServiceDiscovery{
		client:   client,
		services: make(map[string][]string),
		ctx:      ctx,
		cancel:   cancel,
	}

	return sd, nil
}

// DiscoverService 发现指定服务的所有实例
func (sd *ServiceDiscovery) DiscoverService(serviceName string) ([]string, error) {
	key := fmt.Sprintf("/services/%s/", serviceName)
	resp, err := sd.client.Get(sd.ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get service instances: %w", err)
	}

	var addresses []string
	for _, kv := range resp.Kvs {
		addresses = append(addresses, string(kv.Value))
	}

	// 更新本地缓存
	sd.mutex.Lock()
	sd.services[serviceName] = addresses
	sd.mutex.Unlock()

	return addresses, nil
}

// GetService 从缓存中获取服务实例
func (sd *ServiceDiscovery) GetService(serviceName string) []string {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()
	return sd.services[serviceName]
}

// WatchService 监听服务变化
func (sd *ServiceDiscovery) WatchService(serviceName string) {
	key := fmt.Sprintf("/services/%s/", serviceName)
	watchChan := sd.client.Watch(sd.ctx, key, clientv3.WithPrefix())

	go func() {
		for {
			select {
			case <-sd.ctx.Done():
				return
			case watchResp := <-watchChan:
				for _, event := range watchResp.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						log.Printf("Service instance added: %s -> %s", string(event.Kv.Key), string(event.Kv.Value))
					case clientv3.EventTypeDelete:
						log.Printf("Service instance removed: %s", string(event.Kv.Key))
					}
				}
				// 重新发现服务实例
				_, err := sd.DiscoverService(serviceName)
				if err != nil {
					log.Printf("Failed to rediscover service %s: %v", serviceName, err)
				}
			}
		}
	}()
}

// GetServiceAddress 获取服务地址（负载均衡）
func (sd *ServiceDiscovery) GetServiceAddress(serviceName string) (string, error) {
	addresses := sd.GetService(serviceName)
	if len(addresses) == 0 {
		// 尝试重新发现
		var err error
		addresses, err = sd.DiscoverService(serviceName)
		if err != nil {
			return "", fmt.Errorf("failed to discover service %s: %w", serviceName, err)
		}
		if len(addresses) == 0 {
			return "", fmt.Errorf("no available instances for service %s", serviceName)
		}
	}

	// 简单的轮询负载均衡
	// 这里可以实现更复杂的负载均衡算法
	index := int(time.Now().UnixNano()) % len(addresses)
	return addresses[index], nil
}

// Close 关闭服务发现客户端
func (sd *ServiceDiscovery) Close() error {
	sd.cancel()
	return sd.client.Close()
}

// ParseServiceAddress 解析服务地址，提取主机和端口
func ParseServiceAddress(address string) (host, port string, err error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid service address format: %s", address)
	}
	return parts[0], parts[1], nil
}
