package postdog

//go:generate mockgen -source=postdog.go -destination=./mocks/postdog.go

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sync"
)

var (
	// ErrNoTransport means no transport is configured.
	ErrNoTransport = errors.New("no transport")
	// ErrUnconfiguredTransport means a transport with a specific name is not configured.
	ErrUnconfiguredTransport = errors.New("unconfigured transport")
)

// New returns a new *Dog.
func New(opts ...Option) *Dog {
	dog := Dog{transports: make(map[string]Transport)}
	for _, opt := range opts {
		opt(&dog)
	}
	return &dog
}

// Option is a postdog option.
type Option func(*Dog)

// WithTransport configures a transport.
func WithTransport(name string, tr Transport) Option {
	return func(dog *Dog) {
		dog.configureTransport(name, tr)
	}
}

// Dog provides a thread-safe
type Dog struct {
	mux              sync.RWMutex
	transports       map[string]Transport
	defaultTransport string
}

// A Transport is responsible for actually sending mails.
type Transport interface {
	Send(context.Context, Mail) error
}

// A Mail provides the sender, recipients and the mail body as defined in RFC 5322.
type Mail interface {
	// From returns the sender of the mail.
	From() mail.Address
	// Recipients returns the recipients of the mail.
	Recipients() []mail.Address
	// RFC returns the RFC 5322 body / data of the mail.
	RFC() string
}

// Use sets the default transport.
func (dog *Dog) Use(transport string) {
	dog.defaultTransport = transport
}

// Send sends the given mail through the default transport.
// A different transport can be specified by using the `Use()` option:
//		dog.Send(context.TODO(), m, postdog.Use("transport-name"))
func (dog *Dog) Send(ctx context.Context, m Mail, opts ...SendOption) error {
	var cfg sendConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	tr, err := dog.transport(cfg.transport)
	if err != nil {
		return err
	}

	if err = tr.Send(ctx, m); err != nil {
		return fmt.Errorf("transport: %w", err)
	}

	return nil
}

// SendOption is an option for the Send() method of a *Dog.
type SendOption func(*sendConfig)

// Use sets the transport name, that should be used for sending the mail.
func Use(transport string) SendOption {
	return func(cfg *sendConfig) {
		cfg.transport = transport
	}
}

type sendConfig struct {
	transport string
}

// Transport returns either the transport with the given name or an ErrUnconfiguredTransport error.
func (dog *Dog) Transport(name string) (Transport, error) {
	return dog.transport(name)
}

func (dog *Dog) transport(name string) (Transport, error) {
	dog.mux.RLock()
	defer dog.mux.RUnlock()

	if name == "" {
		if dog.defaultTransport != "" {
			return dog.transports[dog.defaultTransport], nil
		}
		return nil, ErrNoTransport
	}

	tr, ok := dog.transports[name]
	if !ok {
		return nil, ErrUnconfiguredTransport
	}

	return tr, nil
}

func (dog *Dog) configureTransport(name string, tr Transport) {
	dog.mux.Lock()
	defer dog.mux.Unlock()
	dog.transports[name] = tr
	if dog.defaultTransport == "" {
		dog.defaultTransport = name
	}
}
