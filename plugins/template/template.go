// Package template provides template support for letter bodies.
package template

import (
	"github.com/bounoable/postdog/office"
)

type plugin struct {
	cfg Config
}

// Plugin ...
func Plugin(opts ...Option) office.Plugin {
	return plugin{
		cfg: newConfig(opts...),
	}
}

func (p plugin) Install(ctx office.PluginContext) {

}
