package gmail

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/gmail/v1"
)

func TestFactory(t *testing.T) {
	tests := []struct {
		name       string
		preRun     func()
		postRun    func()
		config     map[string]interface{}
		wantScopes []string
		wantError  error
	}{
		{
			name: "full config",
			config: map[string]interface{}{
				"credentials": "/path/to/creds.json",
				"scopes":      []string{"scope-a", "scope-b", "scope-c"},
				"jwtSubject":  "subject@example.com",
			},
			wantScopes: []string{"scope-a", "scope-b", "scope-c"},
		},
		{
			name:    "env credentials",
			preRun:  func() { os.Setenv("GMAIL_CREDENTIALS", "/path/to/creds.json") },
			postRun: func() { os.Unsetenv("GMAIL_CREDENTIALS") },
			config: map[string]interface{}{
				"scopes":     []string{"scope-a", "scope-b", "scope-c"},
				"jwtSubject": "subject@example.com",
			},
			wantScopes: []string{"scope-a", "scope-b", "scope-c"},
		},
		{
			name:   "missing credentials",
			preRun: func() { os.Unsetenv("GMAIL_CREDENTIALS") },
			config: map[string]interface{}{
				"scopes":     []string{"scope-a", "scope-b", "scope-c"},
				"jwtSubject": "subject@example.com",
			},
			wantError: ErrNoCredentials,
		},
		{
			name: "default scopes",
			config: map[string]interface{}{
				"credentials": "/path/to/creds.json",
				"jwtSubject":  "subject@example.com",
			},
			wantScopes: []string{gmail.MailGoogleComScope},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.preRun != nil {
				test.preRun()
			}

			tr, err := Factory(context.Background(), test.config)
			assert.True(t, errors.Is(err, test.wantError))

			if test.wantError == nil {
				gmailTransport, ok := tr.(*transport)
				assert.True(t, ok)
				assert.NotNil(t, gmailTransport.newTokenSource)
				assert.Equal(t, test.wantScopes, gmailTransport.scopes)
			}

			if test.postRun != nil {
				test.postRun()
			}
		})
	}
}
