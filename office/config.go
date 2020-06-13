package office

//go:generate mockgen -source=config.go -destination=./mock_office/config.go

import (
	"context"

	"github.com/bounoable/postdog/letter"
)

const (
	// BeforeSend hook is called before a letter is sent.
	BeforeSend = SendHook(iota)
	// AfterSend hook is called after a letter has been sent.
	AfterSend
)

// Config is the office configuration.
type Config struct {
	// Logger is an optional Logger implementation.
	Logger Logger
	// QueueBuffer is the channel buffer size for outgoing letters.
	QueueBuffer int
	// Middleware is the send middleware.
	Middleware []Middleware
	// SendHooks contains the callback functions for the send hooks.
	SendHooks map[SendHook][]func(context.Context, letter.Letter)
	// Plugins are the plugins to used.
	Plugins []Plugin
}

// Middleware intercepts and manipulates letters before they are sent.
// A Middleware that returns an error, aborts the sending of the letter with the returned error.
type Middleware interface {
	Handle(ctx context.Context, let letter.Letter) (letter.Letter, error)
}

// MiddlewareFunc allows the use of functions as middlewares.
type MiddlewareFunc func(ctx context.Context, let letter.Letter) (letter.Letter, error)

// Handle intercepts and manipulates letters before they are sent.
func (fn MiddlewareFunc) Handle(ctx context.Context, let letter.Letter) (letter.Letter, error) {
	return fn(ctx, let)
}

// Logger logs messages.
type Logger interface {
	Log(v ...interface{})
}

// SendHook is a hook point for sending messages.
type SendHook int

// Option is an office option.
type Option func(*Config)

// QueueBuffer sets the outgoing letter channel buffer size.
func QueueBuffer(size int) Option {
	return func(cfg *Config) {
		cfg.QueueBuffer = size
	}
}

// WithMiddleware adds middleware to the office.
func WithMiddleware(middleware ...Middleware) Option {
	return func(cfg *Config) {
		cfg.Middleware = append(cfg.Middleware, middleware...)
	}
}

// WithLogger configures the office to use the given logger.
func WithLogger(logger Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

// WithSendHook registers the given fns to be called at the given SendHook h.
func WithSendHook(h SendHook, fns ...func(context.Context, letter.Letter)) Option {
	return func(cfg *Config) {
		cfg.SendHooks[h] = append(cfg.SendHooks[h], fns...)
	}
}

// WithPlugin installs the given office plugins.
func WithPlugin(plugins ...Plugin) Option {
	return func(cfg *Config) {
		cfg.Plugins = append(cfg.Plugins, plugins...)
	}
}
