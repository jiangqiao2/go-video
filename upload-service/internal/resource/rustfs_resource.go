package resource

import (
	"os"
	"sync"

	"upload-service/pkg/assert"
	"upload-service/pkg/config"
	"upload-service/pkg/logger"
	"upload-service/pkg/manager"
)

var (
	rustfsOnce   sync.Once
	rustfsSingle *RustFSResource
)

type RustFSResource struct {
	endpoint string
	access   string
	secret   string
}

func DefaultRustFSResource() *RustFSResource {
	assert.NotCircular()
	rustfsOnce.Do(func() {
		rustfsSingle = &RustFSResource{}
	})
	assert.NotNil(rustfsSingle)
	return rustfsSingle
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

	logger.Infof("RustFS resource initialized endpoint=%s", endpoint)
}

func (r *RustFSResource) Close() {}

func (r *RustFSResource) GetEndpoint() string  { return r.endpoint }
func (r *RustFSResource) GetAccessKey() string { return r.access }
func (r *RustFSResource) GetSecretKey() string { return r.secret }

type RustFSResourcePlugin struct{}

func (p *RustFSResourcePlugin) Name() string                         { return "rustfsResource" }
func (p *RustFSResourcePlugin) MustCreateResource() manager.Resource { return DefaultRustFSResource() }
