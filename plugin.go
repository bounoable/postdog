package postdog

//go:generate mockgen -source=plugin.go -destination=./mock_postdog/plugin.go

import (
	"context"

	"github.com/bounoable/postdog/letter"
)

// Plugin is an office plugin.
type Plugin interface {
	Install(PluginContext)
}

// PluginFunc allows functions to be used as plugins.
type PluginFunc func(PluginContext)

// Install installs an office plugin.
func (fn PluginFunc) Install(ctx PluginContext) {
	fn(ctx)
}

// PluginContext is the context for a plugin.
// It provides logging and configuration functions.
type PluginContext interface {
	// Log logs messages.
	Log(...interface{})
	// WithSendHook adds callback functions for send hooks.
	WithSendHook(SendHook, ...func(context.Context, letter.Letter))
	// WithMiddleware adds send middleware.
	WithMiddleware(...Middleware)
}

type pluginContext struct {
	cfg *Config
}

func (ctx pluginContext) Log(v ...interface{}) {
	if ctx.cfg.Logger != nil {
		ctx.cfg.Logger.Log(v...)
	}
}

func (ctx pluginContext) WithSendHook(h SendHook, fns ...func(context.Context, letter.Letter)) {
	ctx.cfg.SendHooks[h] = append(ctx.cfg.SendHooks[h], fns...)
}

func (ctx pluginContext) WithMiddleware(mw ...Middleware) {
	ctx.cfg.Middleware = append(ctx.cfg.Middleware, mw...)
}
