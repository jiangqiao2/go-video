package manager

import (
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type (
	ControllerPlugin interface {
		Name() string
		MustCreateController() Controller
	}

	Controller interface {
		RegisterOpenApi(group *gin.RouterGroup)
		RegisterInnerApi(group *gin.RouterGroup)
		RegisterDebugApi(group *gin.RouterGroup)
		RegisterOpsApi(group *gin.RouterGroup)
	}
)

var (
	controllerPlugins = map[string]ControllerPlugin{}
)

// RegisterControllerPlugin registers a controller plugin.
func RegisterControllerPlugin(p ControllerPlugin) {
	if p.Name() == "" {
		panic(fmt.Errorf("%T: empty name", p))
	}
	if existedPlugin, existed := controllerPlugins[p.Name()]; existed {
		panic(fmt.Errorf("%T and %T got same name: %s", p, existedPlugin, p.Name()))
	}
	controllerPlugins[p.Name()] = p
}

// MustInitControllers initialises all registered controllers and attaches routes.
func MustInitControllers(openApiGroup, innerApiGroup, debugApiGroup, opsApiGroup *gin.RouterGroup) {
	for n, p := range controllerPlugins {
		controller := p.MustCreateController()
		if openApiGroup != nil {
			controller.RegisterOpenApi(openApiGroup)
		}
		if innerApiGroup != nil {
			controller.RegisterInnerApi(innerApiGroup)
		}
		if debugApiGroup != nil {
			controller.RegisterDebugApi(debugApiGroup)
		}
		if opsApiGroup != nil {
			controller.RegisterOpsApi(opsApiGroup)
		}
		log.Infof("Register controller: plugin=%s, controller=%+v", n, controller)
	}
}
