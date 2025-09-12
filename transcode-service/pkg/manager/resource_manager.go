package manager

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type (
	ResourcePlugin interface {
		// Name 返回插件的名称，不同 ResourcePlugin 的名称不能相同
		Name() string

		// MustCreateResource 创建 Resource，如果创建失败需要 panic
		MustCreateResource() Resource
	}

	Resource interface {
		// MustOpen 打开资源，如果打开失败需要 panic
		MustOpen()

		// Close 关闭资源
		Close()
	}
)

var (
	resourcePlugins = map[string]ResourcePlugin{}
	resources       []Resource
)

// RegisterResourcePlugin registers resource plugin
func RegisterResourcePlugin(p ResourcePlugin) {
	if p.Name() == "" {
		panic(fmt.Errorf("%T: empty name", p))
	}

	existedPlugin, existed := resourcePlugins[p.Name()]
	if existed {
		panic(fmt.Errorf("%T and %T got same name: %s", p, existedPlugin, p.Name()))
	}

	resourcePlugins[p.Name()] = p
}

// MustInitResources 初始化已注册的 Resource，如果失败则 panic
func MustInitResources() {
	log.Infof("开始初始化资源插件，共有 %d 个插件", len(resourcePlugins))
	for n, p := range resourcePlugins {
		log.Infof("正在初始化资源插件: %s", n)
		resource := p.MustCreateResource()
		log.Infof("资源插件 %s 创建成功，正在打开...", n)
		resource.MustOpen()
		log.Infof("资源插件 %s 打开成功", n)
		resources = append(resources, resource)
		log.Infof("Init resource, plugin=%s, resource=%+v", n, resource)
	}
	log.Infof("所有资源插件初始化完成")
}

// CloseResources 关闭所有注册的资源
func CloseResources() {
	for _, resource := range resources {
		resource.Close()
		log.Infof("Close resource, resource=%+v", resource)
	}
}
