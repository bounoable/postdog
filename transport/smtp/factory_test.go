package smtp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	type wantConfig struct {
		host     string
		port     int
		username string
		password string
	}

	tests := []struct {
		name       string
		config     map[string]interface{}
		wantConfig wantConfig
		wantError  error
	}{
		{
			name: "full config",
			config: map[string]interface{}{
				"host":     "smtp.mailtrap.io",
				"port":     587,
				"username": "user",
				"password": "pass",
			},
			wantConfig: wantConfig{
				host:     "smtp.mailtrap.io",
				port:     587,
				username: "user",
				password: "pass",
			},
		},
		{
			name: "default host = localhost",
			config: map[string]interface{}{
				"port":     25,
				"username": "user",
				"password": "pass",
			},
			wantConfig: wantConfig{
				host:     "localhost",
				port:     25,
				username: "user",
				password: "pass",
			},
		},
		{
			name: "default port = 587",
			config: map[string]interface{}{
				"host":     "smtp.mailtrap.io",
				"username": "user",
				"password": "pass",
			},
			wantConfig: wantConfig{
				host:     "smtp.mailtrap.io",
				port:     587,
				username: "user",
				password: "pass",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr, err := Factory(context.Background(), test.config)
			assert.True(t, errors.Is(err, test.wantError))

			smtpTrans, ok := tr.(*transport)
			assert.True(t, ok)

			if test.wantError == nil {
				assert.Equal(t, test.wantConfig.host, smtpTrans.host)
				assert.Equal(t, test.wantConfig.port, smtpTrans.port)
				assert.Equal(t, test.wantConfig.username, smtpTrans.username)
				assert.Equal(t, test.wantConfig.password, smtpTrans.password)
			}
		})
	}
}
