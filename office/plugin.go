package office

//go:generate mockgen -source=plugin.go -destination=./mock_office/plugin.go

import (
	"context"

	"github.com/bounoable/postdog/letter"
)

// Plugin ...
type Plugin interface {
	Install(PluginContext)
}

// PluginFunc ...
type PluginFunc func(PluginContext)

// Install ...
func (fn PluginFunc) Install(ctx PluginContext) {
	fn(ctx)
}

// PluginContext ...
type PluginContext interface {
	WithSendHook(SendHook, ...func(context.Context, letter.Letter))
}

type pluginContext struct {
	cfg *Config
}

func (ctx pluginContext) WithSendHook(h SendHook, fns ...func(context.Context, letter.Letter)) {
	ctx.cfg.SendHooks[h] = append(ctx.cfg.SendHooks[h], fns...)
}
