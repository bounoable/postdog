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

// WithTransport returns an Option that adds the transport tr with the name in name to a *Dog.
func WithTransport(name string, tr Transport) Option {
	return func(dog *Dog) {
		dog.configureTransport(name, tr)
	}
}

// WithMiddleware returns an Option that adds the middleware mws to a *Dog.
func WithMiddleware(mws ...Middleware) Option {
	return func(dog *Dog) {
		dog.middlewares = append(dog.middlewares, mws...)
	}
}

// A Dog can send mails through one of multiple configured transports.
type Dog struct {
	mux              sync.RWMutex
	transports       map[string]Transport
	defaultTransport string
	middlewares      []Middleware
}

// A Transport is responsible for actually sending mails.
type Transport interface {
	Send(context.Context, Mail) error
}

// Middleware is called on every Send(), allowing manipulation of mails before they are passed to the Transport.
type Middleware interface {
	Handle(context.Context, Mail, func(context.Context, Mail) (Mail, error)) (Mail, error)
}

// A MiddlewareFunc allows functions to be used as Middleware.
type MiddlewareFunc func(context.Context, func(context.Context, Mail) (Mail, error)) (Mail, error)

// Handle calls mw() with the given arguments.
func (mw MiddlewareFunc) Handle(ctx context.Context, fn func(context.Context, Mail) (Mail, error)) (Mail, error) {
	return mw(ctx, fn)
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
	dog.mux.Lock()
	dog.defaultTransport = transport
	dog.mux.Unlock()
}

// Send sends the given mail through the default transport.
//
// A different transport can be specified with the Use() option:
//   dog.Send(context.TODO(), m, postdog.Use("transport-name"))
//
// If the Use() option is used and no transport with the specified name
// has been registered, Send() will return ErrUnconfiguredTransport.
//
// If the Use() option is not used, the default transport will be used instead.
// The default transport is automatically the first transport that has been
// registered and can be overriden by calling dog.Use("transport-name").
// If there's no default transport available, Send() will return ErrNoTransport.
func (dog *Dog) Send(ctx context.Context, m Mail, opts ...SendOption) error {
	var cfg sendConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	tr, err := dog.transport(cfg.transport)
	if err != nil {
		return err
	}

	if m, err = dog.applyMiddleware(ctx, m); err != nil {
		return fmt.Errorf("middleware: %w", err)
	}

	if err = tr.Send(ctx, m); err != nil {
		return fmt.Errorf("transport: %w", err)
	}

	return nil
}

func (dog *Dog) applyMiddleware(ctx context.Context, m Mail) (Mail, error) {
	if len(dog.middlewares) == 0 {
		return m, nil
	}
	return dog.middlewares[0].Handle(ctx, m, dog.nextFunc(0))
}

func (dog *Dog) nextFunc(i int) func(context.Context, Mail) (Mail, error) {
	return func(ctx context.Context, let Mail) (Mail, error) {
		if i >= len(dog.middlewares)-1 {
			return let, nil
		}
		return dog.middlewares[i+1].Handle(ctx, let, dog.nextFunc(i+1))
	}
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
