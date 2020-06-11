package smtp

import (
	"context"

	"github.com/bounoable/postdog/office"
)

const (
	// Provider is the provider name for SMTP transport.
	Provider = "gomail"
)

// AutowireTransport autowires SMTP transport from the given configuration.
func AutowireTransport(_ context.Context, cfg map[string]interface{}) (office.Transport, error) {
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
