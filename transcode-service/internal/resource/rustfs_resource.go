package resource

import (
	"os"
	"sync"

	"transcode-service/pkg/assert"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
)

var (
	rustfsResourceOnce      sync.Once
	singletonRustFSResource *RustFSResource
)

type RustFSResource struct {
	endpoint string
	access   string
	secret   string
}

func DefaultRustFSResource() *RustFSResource {
	assert.NotCircular()
	rustfsResourceOnce.Do(func() {
		singletonRustFSResource = &RustFSResource{}
	})
	assert.NotNil(singletonRustFSResource)
	return singletonRustFSResource
}

func (r *RustFSResource) MustOpen() {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		panic("global config not initialized before RustFSResource")
	}

	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	access := os.Getenv("RUSTFS_ACCESS_KEY")
	secret := os.Getenv("RUSTFS_SECRET_KEY")
	if endpoint == "" {
		endpoint = cfg.RustFS.Endpoint
	}
	if access == "" {
		access = cfg.RustFS.AccessKey
	}
	if secret == "" {
		secret = cfg.RustFS.SecretKey
	}

	if endpoint == "" {
		panic("rustfs endpoint is required")
	}
	if access == "" || secret == "" {
		panic("rustfs access_key and secret_key are required")
	}

	r.endpoint = endpoint
	r.access = access
	r.secret = secret

	logger.Info("RustFS resource initialized", map[string]interface{}{
		"endpoint": endpoint,
	})
}

func (r *RustFSResource) Close() {}

func (r *RustFSResource) GetEndpoint() string  { return r.endpoint }
func (r *RustFSResource) GetAccessKey() string { return r.access }
func (r *RustFSResource) GetSecretKey() string { return r.secret }

type RustFSResourcePlugin struct{}

func (p *RustFSResourcePlugin) Name() string                         { return "rustfsResource" }
func (p *RustFSResourcePlugin) MustCreateResource() manager.Resource { return DefaultRustFSResource() }
