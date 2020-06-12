package office

//go:generate mockgen -source=config.go -destination=./mock_office/config.go

import (
	"context"

	"github.com/bounoable/postdog/letter"
)

// ...
const (
	BeforeSend = SendHook(iota)
	AfterSend
)

// Config is the office configuration.
type Config struct {
	// QueueBuffer is the channel buffer size for outgoing letters.
	QueueBuffer int
	Middleware  []Middleware
	Logger      Logger
	SendHooks   map[SendHook][]func(context.Context, letter.Letter)
	Plugins     []Plugin
}

// Middleware ...
type Middleware interface {
	Handle(ctx context.Context, let letter.Letter) (letter.Letter, error)
}

// MiddlewareFunc ...
type MiddlewareFunc func(ctx context.Context, let letter.Letter) (letter.Letter, error)

// Handle ...
func (fn MiddlewareFunc) Handle(ctx context.Context, let letter.Letter) (letter.Letter, error) {
	return fn(ctx, let)
}

// Logger ...
type Logger interface {
	Log(v ...interface{})
}

// SendHook ...
type SendHook int

// Option ...
type Option func(*Config)

// QueueBuffer ...
func QueueBuffer(size int) Option {
	return func(cfg *Config) {
		cfg.QueueBuffer = size
	}
}

// WithMiddleware ...
func WithMiddleware(middleware ...Middleware) Option {
	return func(cfg *Config) {
		cfg.Middleware = append(cfg.Middleware, middleware...)
	}
}

// WithLogger ...
func WithLogger(logger Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

// WithSendHook ...
func WithSendHook(h SendHook, fns ...func(context.Context, letter.Letter)) Option {
	return func(cfg *Config) {
		cfg.SendHooks[h] = append(cfg.SendHooks[h], fns...)
	}
}

// WithPlugin ...
func WithPlugin(plugins ...Plugin) Option {
	return func(cfg *Config) {
		cfg.Plugins = append(cfg.Plugins, plugins...)
	}
}
