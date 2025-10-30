package registry

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServiceRegistry registers services into etcd.
type ServiceRegistry struct {
	client      *clientv3.Client
	serviceName string
	serviceID   string
	serviceAddr string
	ttl         int64
	leaseID     clientv3.LeaseID
	ctx         context.Context
	cancel      context.CancelFunc
}

// RegistryConfig defines etcd client configuration.
type RegistryConfig struct {
	Endpoints      []string      `yaml:"endpoints"`
	DialTimeout    time.Duration `yaml:"dial_timeout"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
}

// ServiceConfig defines service registration metadata.
type ServiceConfig struct {
	ServiceName     string        `yaml:"service_name"`
	ServiceID       string        `yaml:"service_id"`
	TTL             time.Duration `yaml:"ttl"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

// NewServiceRegistry creates a new ServiceRegistry instance.
func NewServiceRegistry(registryConfig RegistryConfig, serviceConfig ServiceConfig, serviceAddr string) (*ServiceRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   registryConfig.Endpoints,
		DialTimeout: registryConfig.DialTimeout,
		Username:    registryConfig.Username,
		Password:    registryConfig.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ServiceRegistry{
		client:      client,
		serviceName: serviceConfig.ServiceName,
		serviceID:   serviceConfig.ServiceID,
		serviceAddr: serviceAddr,
		ttl:         int64(serviceConfig.TTL.Seconds()),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Register registers service instance.
func (r *ServiceRegistry) Register() error {
	leaseResp, err := r.client.Grant(r.ctx, r.ttl)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	r.leaseID = leaseResp.ID

	key := fmt.Sprintf("/services/%s/%s", r.serviceName, r.serviceID)
	if _, err := r.client.Put(r.ctx, key, r.serviceAddr, clientv3.WithLease(r.leaseID)); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	go r.keepAlive()

	log.Printf("Service registered: %s -> %s", key, r.serviceAddr)
	return nil
}

func (r *ServiceRegistry) keepAlive() {
	ch, err := r.client.KeepAlive(r.ctx, r.leaseID)
	if err != nil {
		log.Printf("Failed to keep alive lease: %v", err)
		return
	}
	for {
		select {
		case <-r.ctx.Done():
			return
		case ka := <-ch:
			if ka == nil {
				log.Println("Keep alive channel closed")
				return
			}
		}
	}
}

// Deregister removes service registration.
func (r *ServiceRegistry) Deregister() error {
	r.cancel()
	if r.leaseID != 0 {
		if _, err := r.client.Revoke(context.Background(), r.leaseID); err != nil {
			log.Printf("Failed to revoke lease: %v", err)
		}
	}
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close etcd client: %w", err)
	}
	log.Printf("Service deregistered: %s", r.serviceID)
	return nil
}
