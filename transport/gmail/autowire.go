package gmail

import (
	"context"

	"github.com/bounoable/postdog/office"
)

const (
	// Provider is the provider name for Gmail.
	Provider = "gmail"
)

// AutowireTransport autowires Gmail transport from the given config.
func AutowireTransport(ctx context.Context, cfg map[string]interface{}) (office.Transport, error) {
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
