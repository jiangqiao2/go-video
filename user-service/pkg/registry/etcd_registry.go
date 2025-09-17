package registry

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServiceRegistry etcd服务注册器
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

// RegistryConfig 注册配置
type RegistryConfig struct {
	Endpoints      []string      `yaml:"endpoints"`
	DialTimeout    time.Duration `yaml:"dial_timeout"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServiceName     string        `yaml:"service_name"`
	ServiceID       string        `yaml:"service_id"`
	TTL             time.Duration `yaml:"ttl"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

// NewServiceRegistry 创建服务注册器
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

// Register 注册服务
func (r *ServiceRegistry) Register() error {
	// 创建租约
	leaseResp, err := r.client.Grant(r.ctx, r.ttl)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	r.leaseID = leaseResp.ID

	// 注册服务
	key := fmt.Sprintf("/services/%s/%s", r.serviceName, r.serviceID)
	_, err = r.client.Put(r.ctx, key, r.serviceAddr, clientv3.WithLease(r.leaseID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 启动续约
	go r.keepAlive()

	log.Printf("Service registered: %s -> %s", key, r.serviceAddr)
	return nil
}

// keepAlive 保持租约活跃
func (r *ServiceRegistry) keepAlive() {
	ch, kaerr := r.client.KeepAlive(r.ctx, r.leaseID)
	if kaerr != nil {
		log.Printf("Failed to keep alive lease: %v", kaerr)
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
			// log.Printf("Lease %d renewed", ka.ID)
		}
	}
}

// Deregister 注销服务
func (r *ServiceRegistry) Deregister() error {
	r.cancel()

	// 撤销租约
	if r.leaseID != 0 {
		_, err := r.client.Revoke(context.Background(), r.leaseID)
		if err != nil {
			log.Printf("Failed to revoke lease: %v", err)
		}
	}

	// 关闭客户端
	err := r.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close etcd client: %w", err)
	}

	log.Printf("Service deregistered: %s", r.serviceID)
	return nil
}
