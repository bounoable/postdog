package gmail

//go:generate mockgen -source=gmail.go -destination=./mocks/gmail.go

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/bounoable/postdog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var (
	// ErrNoCredentials means no credentials are provided to initialize the Gmail service.
	ErrNoCredentials = errors.New("no credentials provided")
)

// Transport returns a Gmail transport with the `gmail.GmailGoogleComScope` (https://mail.google.com/) scope.
//
// Providing credentials
//
// If the `GMAIL_CREDENTIALS` environment variable is not empty, the file at
// the given filepath will be used as the credentials file. Otherwise credentials
// must be provided with one of the CredentialsXXX() options.
//
// Gmail service setup happens with the first call to Send(), if WithSender()
// option is not used. If the setup fails, Send() will return a *SetupError.
// If the setup fails due missing credentials, the error will unwrap to an
// ErrNoCredentials error.
//
// Authentication will happen with the first call to Send(). If authentication
// fails, Send() will return a *SetupError.
//
// Specifying scopes
//
// Scopes may be specified with the Scopes() option. If no scopes are specified,
// only the `gmail.GmailGoogleComScope` will be used. If it' used with a non-nil
// empty slice, no scopes will be used.
func Transport(opts ...Option) postdog.Transport {
	if credsPath := os.Getenv("GMAIL_CREDENTIALS"); credsPath != "" {
		opts = append([]Option{CredentialsFile(credsPath)}, opts...)
	}

	t := transport{newSender: newGmailSender}
	for _, opt := range opts {
		opt(&t)
	}

	// check for nil and not for length, so users can remove all scopes
	if t.scopes == nil {
		t.scopes = []string{gmail.MailGoogleComScope}
	}

	return &t
}

// Option is an option for the Gmail transport.
type Option func(*transport)

type transport struct {
	sync.RWMutex
	scopes         []string
	clientOpts     []option.ClientOption
	sender         Sender
	newSender      func(context.Context, oauth2.TokenSource, ...option.ClientOption) (Sender, error)
	tokenSource    oauth2.TokenSource
	newTokenSource func(context.Context, ...string) (oauth2.TokenSource, error)
}

// Sender wraps the *gmail.UsersMessagesService.Send().Do() method(s).
type Sender interface {
	Send(userID string, msg *gmail.Message) error
}

// WithSender returns an Option that sets the Sender for sending mails.
//
// Using this option makes the following options no-ops:
// WithSenderFactory(), WithClientOptions(), WithTokenSource(),
// WithTokenSourceFactory(), CredentialsXXX(), JWTSubject()
func WithSender(s Sender) Option {
	return func(t *transport) {
		t.sender = s
	}
}

// WithSenderFactory return an Option that sets the returned Sender of newSender as the Sender for sending mails.
//
// Using this option makes the following options no-ops:
// WithClientOptions(), WithTokenSource(), WithTokenSourceFactory(),
// CredentialsXXX(), JWTSubject()
func WithSenderFactory(
	newSender func(
		context.Context,
		oauth2.TokenSource,
		...option.ClientOption,
	) (Sender, error),
) Option {
	return func(t *transport) {
		t.newSender = newSender
	}
}

// WithClientOptions returns an Option that adds custom option.ClientOptions that will be passed to the Sender factory.
func WithClientOptions(opts ...option.ClientOption) Option {
	return func(t *transport) {
		t.clientOpts = append(t.clientOpts, opts...)
	}
}

// WithTokenSource returns an Option that sets ts as the oauth2.TokenSource
// to be passed as an option.ClientOption to the Sender factory.
//
// Using this option makes the following options no-ops:
// WithTokenSourceFactory(), CredentialsXXX(), JWTSubject()
func WithTokenSource(ts oauth2.TokenSource) Option {
	return func(t *transport) {
		t.tokenSource = ts
	}
}

// WithTokenSourceFactory returns an Option that sets newTokenSource to be used as the oauth2.TokenSource factory.
//
// Using this option makes the following options no-ops:
// CredentialsXXX(), JWTSubject()
func WithTokenSourceFactory(newTokenSource func(context.Context, ...string) (oauth2.TokenSource, error)) Option {
	return func(t *transport) {
		t.newTokenSource = newTokenSource
	}
}

// Scopes returns an Option that specifies the Gmail scopes.
func Scopes(scopes ...string) Option {
	return func(t *transport) {
		t.scopes = append(t.scopes, scopes...)
	}
}

// Credentials returns an Option that authenticates API calls using creds.
func Credentials(creds *google.Credentials) Option {
	return WithTokenSourceFactory(func(context.Context, ...string) (oauth2.TokenSource, error) {
		if creds == nil {
			return nil, errors.New("nil credentials")
		}
		return creds.TokenSource, nil
	})
}

// CredentialsJSON returns an Option that authenticates API calls using the credentials file in jsonKey.
//
// Use the JWTSubject() option to specify the `subject` field of the parsed *jwt.Config.
func CredentialsJSON(jsonKey []byte, opts ...JWTConfigOption) Option {
	return WithTokenSourceFactory(func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
		cfg, err := google.JWTConfigFromJSON(jsonKey, scopes...)
		if err != nil {
			return nil, fmt.Errorf("parse credentials: %w", err)
		}
		for _, opt := range opts {
			opt(cfg)
		}
		return cfg.TokenSource(ctx), nil
	})
}

// CredentialsFile returns an Option that authenticates API calls using the credentials file at path.
//
// Use the JWTSubject() option to specify the `subject` field of the parsed JWT config.
func CredentialsFile(path string, opts ...JWTConfigOption) Option {
	return WithTokenSourceFactory(func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read credentials file %s: %w", path, err)
		}
		cfg, err := google.JWTConfigFromJSON(b, scopes...)
		if err != nil {
			return nil, fmt.Errorf("parse credentials: %w", err)
		}
		for _, opt := range opts {
			opt(cfg)
		}
		return cfg.TokenSource(ctx), nil
	})
}

// JWTConfigOption configures a *jwt.Config.
type JWTConfigOption func(*jwt.Config)

// JWTSubject returns an Option that sets the `subject` field of the JWT config.
func JWTSubject(subject string) JWTConfigOption {
	return func(cfg *jwt.Config) {
		cfg.Subject = subject
	}
}

func (tr *transport) Send(ctx context.Context, m postdog.Mail) error {
	if err := tr.ensure(ctx); err != nil {
		return err
	}

	if err := tr.sender.Send("me", &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(m.RFC())),
	}); err != nil {
		return fmt.Errorf("gmail: %w", err)
	}

	return nil
}

func (tr *transport) ensure(ctx context.Context) error {
	if tr.initialized() {
		return nil
	}

	if err := tr.init(ctx); err != nil {
		return &SetupError{err}
	}

	return nil
}

func (tr *transport) initialized() bool {
	tr.RLock()
	defer tr.RUnlock()
	return tr.sender != nil
}

func (tr *transport) init(ctx context.Context) error {
	tr.Lock()
	defer tr.Unlock()

	if tr.tokenSource == nil && tr.newTokenSource != nil {
		ts, err := tr.newTokenSource(ctx)
		if err != nil {
			return fmt.Errorf("new token source: %w", err)
		}
		tr.tokenSource = ts
	}

	if tr.tokenSource != nil {
		s, err := tr.newSender(ctx, tr.tokenSource, tr.clientOpts...)
		if err != nil {
			return fmt.Errorf("new sender: %w", err)
		}
		tr.sender = s
	}

	if tr.sender == nil {
		return ErrNoCredentials
	}

	return nil
}

func newGmailSender(ctx context.Context, ts oauth2.TokenSource, opts ...option.ClientOption) (Sender, error) {
	opts = append([]option.ClientOption{option.WithTokenSource(ts)}, opts...)
	svc, err := gmail.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create gmail service: %w", err)
	}
	return &gmailSender{svc}, nil
}

type gmailSender struct {
	svc *gmail.Service
}

func (s *gmailSender) Send(userID string, msg *gmail.Message) error {
	_, err := s.svc.Users.Messages.Send(userID, msg).Do()
	return err
}

// SetupError is an error that occurred during the initial setup of the Gmail service.
type SetupError struct {
	Err error
}

func (err SetupError) Unwrap() error {
	return err.Err
}

func (err SetupError) Error() string {
	return fmt.Sprintf("setup: %s", err.Err.Error())
}
