package smtp

import (
	"context"

	"github.com/bounoable/postdog"
)

// Factory accepts configuration as a map[string]interface{} and instantiates the SMTP transport from it.
//
// Example configuration:
//   cfg := map[string]interface{}{
//     "host": "smtp.mailtrap.io",
//     "port": 587,
//     "username": "abcdef123456",
//     "password": "654321fedcba",
//   }
//
// Default host is "localhost". Default port is 587.
func Factory(_ context.Context, cfg map[string]interface{}) (postdog.Transport, error) {
	host, ok := cfg["host"].(string)
	if !ok {
		host = "localhost"
	}
	port, ok := cfg["port"].(int)
	if !ok {
		port = 587
	}
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)

	return Transport(host, port, username, password), nil
}
