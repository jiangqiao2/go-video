package cache

import (
	"user-service/internal/resource"
	"user-service/pkg/manager"
	"user-service/pkg/revocation"
)

type revocationComponent struct{}

func (c *revocationComponent) Start() error {
	if revocation.DefaultRevocationStore() == nil {
		revocation.Init(NewRedisRevocationStore(resource.DefaultRedisResource().Client()))
	}
	return nil
}

func (c *revocationComponent) Stop() error     { return nil }
func (c *revocationComponent) GetName() string { return "revocation" }

type revocationComponentPlugin struct{}

func (p *revocationComponentPlugin) Name() string { return "revocation" }
func (p *revocationComponentPlugin) MustCreateComponent(deps *manager.Dependencies) manager.Component {
	return &revocationComponent{}
}

func init() {
	manager.RegisterComponentPlugin(&revocationComponentPlugin{})
}
