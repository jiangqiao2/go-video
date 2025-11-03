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

// ServiceDiscovery provides simple etcd-based service discovery.
type ServiceDiscovery struct {
	client   *clientv3.Client
	services map[string][]string
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewServiceDiscovery initialises discovery client with registry config.
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
	return &ServiceDiscovery{
		client:   client,
		services: make(map[string][]string),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// DiscoverService fetches available instances from etcd and caches them.
func (sd *ServiceDiscovery) DiscoverService(serviceName string) ([]string, error) {
	key := fmt.Sprintf("/services/%s/", serviceName)
	resp, err := sd.client.Get(sd.ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get service instances: %w", err)
	}

	addresses := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		addresses = append(addresses, string(kv.Value))
	}

	sd.mutex.Lock()
	sd.services[serviceName] = addresses
	sd.mutex.Unlock()

	return addresses, nil
}

// GetService returns cached instances if available.
func (sd *ServiceDiscovery) GetService(serviceName string) []string {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()
	return sd.services[serviceName]
}

// WatchService subscribes to etcd updates for the target service.
func (sd *ServiceDiscovery) WatchService(serviceName string) {
	key := fmt.Sprintf("/services/%s/", serviceName)
	watchCh := sd.client.Watch(sd.ctx, key, clientv3.WithPrefix())

	go func() {
		for {
			select {
			case <-sd.ctx.Done():
				return
			case resp := <-watchCh:
				for _, event := range resp.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						log.Printf("Service instance added: %s -> %s", string(event.Kv.Key), string(event.Kv.Value))
					case clientv3.EventTypeDelete:
						log.Printf("Service instance removed: %s", string(event.Kv.Key))
					}
				}
				if _, err := sd.DiscoverService(serviceName); err != nil {
					log.Printf("Failed to refresh service cache for %s: %v", serviceName, err)
				}
			}
		}
	}()
}

// GetServiceAddress returns one instance using naive round-robin.
func (sd *ServiceDiscovery) GetServiceAddress(serviceName string) (string, error) {
	addresses := sd.GetService(serviceName)
	if len(addresses) == 0 {
		var err error
		addresses, err = sd.DiscoverService(serviceName)
		if err != nil {
			return "", fmt.Errorf("failed to discover service %s: %w", serviceName, err)
		}
		if len(addresses) == 0 {
			return "", fmt.Errorf("no available instances for service %s", serviceName)
		}
	}
	idx := int(time.Now().UnixNano()) % len(addresses)
	return addresses[idx], nil
}

// Close releases resources and stops watchers.
func (sd *ServiceDiscovery) Close() error {
	sd.cancel()
	return sd.client.Close()
}

// ParseServiceAddress splits host:port formatted address.
func ParseServiceAddress(address string) (string, string, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid service address format: %s", address)
	}
	return parts[0], parts[1], nil
}
