package send

import "time"

// Option is a send option.
type Option func(*Config)

// Config is the send config.
type Config struct {
	Transport string
	Timeout   time.Duration
}

// Use sets the transport name that should be used for sending a Mail.
func Use(transport string) Option {
	return func(cfg *Config) {
		cfg.Transport = transport
	}
}

// Timeout returns an Option that adds a timeout a send.
func Timeout(dur time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = dur
	}
}
