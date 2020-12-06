package dispatch

import "github.com/bounoable/postdog"

// Config is the dispatch config.
type Config struct {
	SendOptions []postdog.SendOption
}

// Option is a dispatch option.
type Option func(*Config)

// Configure builds Config from opts.
func Configure(opts ...Option) Config {
	var cfg Config
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// SendOptions returns an Option that adds postdog.SendOptions to the dispatch.
func SendOptions(opts ...postdog.SendOption) Option {
	return func(cfg *Config) {
		cfg.SendOptions = append(cfg.SendOptions, opts...)
	}
}
