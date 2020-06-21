package template

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
)

// Name is the plugin name.
var Name = "template"

// Register registers the plugin factory in the autowire config.
func Register(cfg *autowire.Config) {
	cfg.RegisterPlugin(Name, autowire.PluginFactoryFunc(AutowirePlugin))
}

// AutowirePlugin creates the templat plugin from the given cfg.
func AutowirePlugin(_ context.Context, cfg map[string]interface{}) (postdog.Plugin, error) {
	dirs, _ := cfg["dirs"].([]string)
	templates, _ := cfg["templates"].(map[string]interface{})

	var opts []Option
	for _, dir := range dirs {
		opts = append(opts, UseDir(dir))
	}

	for name, path := range templates {
		spath, ok := path.(string)
		if !ok {
			continue
		}
		opts = append(opts, Use(name, spath))
	}

	return TryPlugin(opts...)
}
