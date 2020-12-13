package dispatch

import (
	"time"

	"github.com/bounoable/postdog/send"
)

// Config is the dispatch config.
type Config struct {
	Send    send.Config
	Timeout time.Duration

	sendOpts []send.Option
}

// Option is a dispatch option.
type Option func(*Config)

// Configure builds Config from opts.
func Configure(opts ...Option) Config {
	var cfg Config
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.Send = send.Configure(cfg.sendOpts...)
	cfg.sendOpts = nil
	return cfg
}

// SendOptions returns an Option that adds send.Options to the dispatch.
func SendOptions(opts ...send.Option) Option {
	return func(cfg *Config) {
		cfg.sendOpts = append(cfg.sendOpts, opts...)
	}
}

// Timeout returns an Option that adds a timeout to the dispatch.
// The timeout applies only to the dispatch, not to the actual sending of the mail.
func Timeout(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = d
	}
}
