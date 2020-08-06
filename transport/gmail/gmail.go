// Package gmail provides the transport implementation for gmail.
package gmail

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type transport struct {
	scopes        []string
	clientOptions []option.ClientOption
	svc           *gmail.Service
}

// NewTransport initializes a gmail transport.
// Credentials must be provided, if GMAIL_CREDENTIALS environment variable is not set.
// If no scopes are provided, the gmail.MailGoogleComScope (https://mail.google.com/) scope is used.
func NewTransport(ctx context.Context, options ...Option) (postdog.Transport, error) {
	if credsPath := os.Getenv("GMAIL_CREDENTIALS"); credsPath != "" {
		options = append([]Option{CredentialsFile(credsPath)}, options...)
	}

	var trans transport
	for _, opt := range options {
		opt(&trans)
	}

	if len(trans.scopes) == 0 {
		trans.scopes = []string{gmail.MailGoogleComScope}
	}

	if err := trans.createService(ctx); err != nil {
		return nil, err
	}

	return trans, nil
}

// Option is a gmail transport option.
type Option func(*transport)

// Scopes adds gmail scopes.
func Scopes(scopes ...string) Option {
	return func(trans *transport) {
		trans.scopes = scopes
	}
}

// Credentials adds an option.ClientOption that authenticates API calls.
func Credentials(creds *google.Credentials) Option {
	return func(trans *transport) {
		ClientOptions(option.WithCredentials(creds))(trans)
	}
}

// CredentialsJSON adds an option.ClientOption that authenticates API calls with the given service account or refresh token JSON credentials.
func CredentialsJSON(p []byte) Option {
	return func(trans *transport) {
		ClientOptions(option.WithCredentialsJSON(p))(trans)
	}
}

// CredentialsFile adds an option.ClientOption that authenticates API calls with the given service account or refresh token JSON credentials file.
func CredentialsFile(filename string) Option {
	return func(trans *transport) {
		ClientOptions(option.WithCredentialsFile(filename))(trans)
	}
}

// ClientOptions adds custom option.ClientOption to the gmail service.
func ClientOptions(options ...option.ClientOption) Option {
	return func(trans *transport) {
		trans.clientOptions = append(trans.clientOptions, options...)
	}
}

func (trans *transport) createService(ctx context.Context) error {
	svc, err := gmail.NewService(ctx, trans.clientOptions...)
	if err != nil {
		return err
	}
	trans.svc = svc
	return nil
}

func (trans transport) Send(ctx context.Context, let letter.Letter) error {
	msg := let.RFC()
	gmsg := gmail.Message{Raw: base64.RawURLEncoding.EncodeToString([]byte(msg))}
	if _, err := trans.svc.Users.Messages.Send(let.From.Address, &gmsg).Do(); err != nil {
		return err
	}
	return nil
}
