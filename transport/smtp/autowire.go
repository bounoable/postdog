package smtp

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
)

const (
	// Provider is the provider name for SMTP transport.
	Provider = "smtp"
)

// Register registers the transport factory in the autowire config.
func Register(cfg *autowire.Config) {
	cfg.RegisterProvider(Provider, autowire.TransportFactoryFunc(AutowireTransport))
}

// AutowireTransport autowires SMTP transport from the given configuration.
func AutowireTransport(_ context.Context, cfg map[string]interface{}) (postdog.Transport, error) {
	host, ok := cfg["host"].(string)
	if !ok {
		host = ""
	}

	port, ok := cfg["port"].(int)
	if !ok {
		port = 587
	}

	username, ok := cfg["username"].(string)
	if !ok {
		username = ""
	}

	password, ok := cfg["password"].(string)
	if !ok {
		host = ""
	}

	return NewTransport(host, port, username, password), nil
}
