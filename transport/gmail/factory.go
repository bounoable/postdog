package gmail

import (
	"context"
	"os"

	"github.com/bounoable/postdog"
)

// Factory accepts configuration as a map[string]interface{} and instantiates the Gmail transport from it.
//
// Example configuration:
//   cfg := map[string]interface{}{
//     "credentials": "/path/to/credentials.json",
//     "scopes": []string{gmail.MailGoogleComScope},
//     "jwtSubject": "bob@example.com",
//   }
func Factory(_ context.Context, cfg map[string]interface{}) (postdog.Transport, error) {
	var opts []Option
	var jwtOpts []JWTConfigOption

	if scopes, ok := cfg["scopes"].([]string); ok {
		opts = append(opts, Scopes(scopes...))
	}

	if jwtSubject, ok := cfg["jwtSubject"].(string); ok {
		jwtOpts = append(jwtOpts, JWTSubject(jwtSubject))
	}

	credentials, ok := cfg["credentials"].(string)
	if !ok {
		if creds := os.Getenv("GMAIL_CREDENTIALS"); creds == "" {
			return nil, ErrNoCredentials
		}
	}
	opts = append(opts, CredentialsFile(credentials, jwtOpts...))

	return Transport(opts...), nil
}
