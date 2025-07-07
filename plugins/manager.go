package plugins

import "finishy1995/mongo-adapter/library/log"

var (
	pluginManager = map[string]Plugin{}
)

func RegisterPlugin(plugin Plugin) {
	pluginManager[plugin.Name()] = plugin
}

func StartPlugins(plugins []string) {
	for _, plugin := range plugins {
		if p, ok := pluginManager[plugin]; ok {
			if err := p.Init(); err != nil {
				log.Errorf("[%s] init failed, err: %v", plugin, err)
			} else {
				log.Infof("[%s] init success", plugin)
			}
		} else {
			log.Errorf("[%s] not found", plugin)
		}
	}
}
