// Package gmail provides the transport implementation for gmail.
package gmail

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type transport struct {
	scopes         []string
	clientOptions  []option.ClientOption
	getTokenSource func(context.Context) (oauth2.TokenSource, error)
	mux            sync.RWMutex
	tokenSource    oauth2.TokenSource
	svc            *gmail.Service
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

	if trans.getTokenSource == nil {
		trans.getTokenSource = trans.getDefaultTokenSource
	}

	if err := trans.createService(ctx); err != nil {
		return nil, fmt.Errorf("create gmail service: %w", err)
	}

	return &trans, nil
}

// Option is a gmail transport option.
type Option func(*transport)

// Scopes adds gmail scopes.
func Scopes(scopes ...string) Option {
	return func(trans *transport) {
		trans.scopes = scopes
	}
}

// Credentials authenticates API calls with the creds.TokenSource.
func Credentials(creds *google.Credentials) Option {
	return func(trans *transport) {
		trans.getTokenSource = func(_ context.Context) (oauth2.TokenSource, error) {
			if creds == nil {
				return nil, errors.New("nil credentials")
			}

			return creds.TokenSource, nil
		}
	}
}

// CredentialsJSON authenticates API calls with the JSON credentials in p.
func CredentialsJSON(p []byte, opts ...JWTConfigOption) Option {
	return func(trans *transport) {
		trans.getTokenSource = func(ctx context.Context) (oauth2.TokenSource, error) {
			cfg, err := google.JWTConfigFromJSON(p, trans.scopes...)
			if err != nil {
				return nil, fmt.Errorf("parse credentials: %w", err)
			}

			for _, opt := range opts {
				opt(cfg)
			}

			return cfg.TokenSource(ctx), nil
		}
	}
}

// JWTConfigOption ...
type JWTConfigOption func(*jwt.Config)

// JWTSubject ...
func JWTSubject(subject string) JWTConfigOption {
	return func(cfg *jwt.Config) {
		cfg.Subject = subject
	}
}

// CredentialsFile authenticates API calls with the given credentials file.
func CredentialsFile(filename string, opts ...JWTConfigOption) Option {
	return func(trans *transport) {
		trans.getTokenSource = func(ctx context.Context) (oauth2.TokenSource, error) {
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, fmt.Errorf("read credentials file: %w", err)
			}

			cfg, err := google.JWTConfigFromJSON(b, trans.scopes...)
			if err != nil {
				return nil, fmt.Errorf("parse credentials: %w", err)
			}

			for _, opt := range opts {
				opt(cfg)
			}

			return cfg.TokenSource(ctx), nil
		}
	}
}

// ClientOptions adds custom option.ClientOption to the gmail service.
func ClientOptions(options ...option.ClientOption) Option {
	return func(trans *transport) {
		trans.clientOptions = append(trans.clientOptions, options...)
	}
}

func (trans *transport) createService(ctx context.Context) error {
	ts, err := trans.getCachedTokenSource(ctx)

	svc, err := gmail.NewService(ctx, append(
		[]option.ClientOption{option.WithTokenSource(ts)},
		trans.clientOptions...,
	)...)

	if err != nil {
		return err
	}
	trans.svc = svc

	return nil
}

func (trans *transport) getCachedTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	trans.mux.RLock()
	if trans.tokenSource != nil {
		trans.mux.RUnlock()
		return trans.tokenSource, nil
	}
	trans.mux.RUnlock()
	trans.mux.Lock()
	defer trans.mux.Unlock()

	ts, err := trans.getTokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token source: %w", err)
	}
	trans.tokenSource = ts
	return trans.tokenSource, nil
}

func (trans *transport) getDefaultTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	ts, err := google.DefaultTokenSource(ctx, trans.scopes...)
	if err != nil {
		return nil, fmt.Errorf("get default token source: %w", err)
	}
	return ts, nil
}

func (trans *transport) Send(ctx context.Context, let letter.Letter) error {
	msg := let.RFC()
	gmsg := gmail.Message{Raw: base64.URLEncoding.EncodeToString([]byte(msg))}
	if _, err := trans.svc.Users.Messages.Send("me", &gmsg).Do(); err != nil {
		return err
	}
	return nil
}
