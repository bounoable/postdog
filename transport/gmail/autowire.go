package gmail

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
)

const (
	// Provider is the provider name for Gmail.
	Provider = "gmail"
)

func init() {
	autowire.RegisterProvider(Provider, autowire.TransportFactoryFunc(AutowireTransport))
}

// Register registers the transport factory in the autowire config.
func Register(cfg *autowire.Config) {
	cfg.RegisterProvider(Provider, autowire.TransportFactoryFunc(AutowireTransport))
}

// AutowireTransport autowires gmail transport from the given config.
func AutowireTransport(ctx context.Context, cfg map[string]interface{}) (postdog.Transport, error) {
	var sscopes []string
	if scopes, ok := cfg["scopes"].([]interface{}); ok {
		for _, scope := range scopes {
			if s, ok := scope.(string); ok {
				sscopes = append(sscopes, s)
			}
		}
	}

	serviceAccountPath, ok := cfg["serviceAccount"].(string)
	if !ok {
		serviceAccountPath = ""
	}

	return NewTransport(ctx, []Option{Scopes(sscopes...), CredentialsFile(serviceAccountPath)}...)
}
