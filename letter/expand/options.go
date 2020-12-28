package expand

import "github.com/bounoable/postdog/letter/rfc"

// Config is the expand config.
type Config struct {
	RFCOptions []rfc.Option
}

// Option is an expand option.
type Option func(*Config)

// Configure makes a Config from opts.
func Configure(opts ...Option) (cfg Config) {
	for _, opt := range opts {
		opt(&cfg)
	}
	return
}

// RFCOptions returns an Option that adds rfc.Options when generating the RFC body.
func RFCOptions(opts ...rfc.Option) Option {
	return func(cfg *Config) {
		cfg.RFCOptions = append(cfg.RFCOptions, opts...)
	}
}
