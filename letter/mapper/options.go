package mapper

// Config is the mapper config.
type Config struct {
	WithoutAttachmentContent bool
}

// Option configures a Map() call.
type Option func(*Config)

// Configure makes a Config from opts.
func Configure(opts ...Option) (cfg Config) {
	for _, opt := range opts {
		opt(&cfg)
	}
	return
}

// WithoutAttachmentContent returns a MapOption that clears the content of one or multiple Attachments.
func WithoutAttachmentContent() Option {
	return func(cfg *Config) {
		cfg.WithoutAttachmentContent = true
	}
}
