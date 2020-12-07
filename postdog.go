package postdog

//go:generate mockgen -source=postdog.go -destination=./mocks/postdog.go

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sync"
	"time"
)

const (
	// BeforeSend is the Hook that's called before a mail is sent.
	BeforeSend = Hook(iota + 1)
	// AfterSend is the Hook that's called after a mail has been sent.
	AfterSend
)

var (
	// ErrNoTransport means no transport is configured.
	ErrNoTransport = errors.New("no transport")
	// ErrUnconfiguredTransport means a transport with a specific name is not configured.
	ErrUnconfiguredTransport = errors.New("unconfigured transport")
)

// A Dog can send mails through one of multiple configured transports.
type Dog struct {
	mux              sync.RWMutex
	transports       map[string]Transport
	defaultTransport string
	middlewares      []Middleware
	hooks            map[Hook][]Listener
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
type MiddlewareFunc func(context.Context, Mail, func(context.Context, Mail) (Mail, error)) (Mail, error)

// Option is a *Dog option.
type Option interface {
	Apply(*Dog)
}

// OptionFunc allows functions to be used as Options.
type OptionFunc func(*Dog)

// A Plugin is a collection of Options.
type Plugin []Option

// A Mail provides the sender, recipients and the mail body as defined in RFC 5322.
type Mail interface {
	// From returns the sender of the mail.
	From() mail.Address
	// Recipients returns the recipients of the mail.
	Recipients() []mail.Address
	// RFC returns the RFC 5322 body / data of the mail.
	RFC() string
}

// A Waiter implements rate limiting.
type Waiter interface {
	// Wait should block until the next mail can be sent.
	Wait(context.Context) error
}

// SendOption is an option for the Send() method of a *Dog.
type SendOption func(*sendConfig)

type sendConfig struct {
	transport string
	timeout   time.Duration
}

// A Hook is a hook point.
type Hook uint8

// Listener is a callback for a Hook.
type Listener interface {
	Handle(context.Context, Hook, Mail)
}

// ListenerFunc allows functions to be used as Listeners.
type ListenerFunc func(context.Context, Hook, Mail)

type ctxKey string

// New returns a new *Dog.
func New(opts ...Option) *Dog {
	dog := Dog{
		transports: make(map[string]Transport),
		hooks:      make(map[Hook][]Listener),
	}
	for _, opt := range opts {
		opt.Apply(&dog)
	}
	return &dog
}

// WithTransport returns an OptionFunc that adds the transport tr with the name in name to a *Dog.
func WithTransport(name string, tr Transport) OptionFunc {
	return func(dog *Dog) {
		dog.configureTransport(name, tr)
	}
}

// WithMiddleware returns an OptionFunc that adds the middleware mws to a *Dog.
func WithMiddleware(mws ...Middleware) OptionFunc {
	return func(dog *Dog) {
		dog.middlewares = append(dog.middlewares, mws...)
	}
}

// WithMiddlewareFunc returns an OptionFunc that adds the middleware mws to a *Dog.
func WithMiddlewareFunc(mws ...MiddlewareFunc) OptionFunc {
	mw := make([]Middleware, len(mws))
	for i, m := range mws {
		mw[i] = Middleware(m)
	}
	return WithMiddleware(mw...)
}

// WithRateLimiter returns an OptionFunc that adds a middleware to a *Dog.
//
// The middleware will call rl.Wait() for every mail that's sent.
func WithRateLimiter(rl Waiter) OptionFunc {
	return WithMiddlewareFunc(func(
		ctx context.Context,
		m Mail,
		next func(context.Context, Mail) (Mail, error),
	) (Mail, error) {
		if err := rl.Wait(ctx); err != nil {
			return m, fmt.Errorf("rate limiter: %w", err)
		}
		return next(ctx, m)
	})
}

// WithHook returns an OptionFunc that adds Listener l for Hook h to a *Dog.
func WithHook(h Hook, l Listener) OptionFunc {
	return func(dog *Dog) {
		dog.hooks[h] = append(dog.hooks[h], l)
	}
}

// Use sets the transport name that should be used for sending a Mail.
func Use(transport string) SendOption {
	return func(cfg *sendConfig) {
		cfg.transport = transport
	}
}

// Timeout returns an Option that adds a timeout a send.
func Timeout(dur time.Duration) SendOption {
	return func(cfg *sendConfig) {
		cfg.timeout = dur
	}
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

	var cancel context.CancelFunc
	if cfg.timeout == 0 {
		ctx, cancel = context.WithCancel(ctx)
	} else {
		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
	}
	defer cancel()

	tr, err := dog.transport(cfg.transport)
	if err != nil {
		return err
	}

	if m, err = dog.applyMiddleware(ctx, m); err != nil {
		return fmt.Errorf("middleware: %w", err)
	}

	dog.callHooks(ctx, BeforeSend, m)
	defer func() { dog.callHooks(ctx, AfterSend, m) }()

	err = tr.Send(ctx, m)
	if err != nil {
		ctx = withSendError(ctx, err)
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

func (dog *Dog) callHooks(ctx context.Context, h Hook, m Mail) {
	for _, lis := range dog.listeners(h) {
		go lis.Handle(ctx, h, m)
	}
}

func (dog *Dog) listeners(h Hook) []Listener {
	dog.mux.RLock()
	defer dog.mux.RUnlock()
	return dog.hooks[h]
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

// Handle calls mw() with the given arguments.
func (mw MiddlewareFunc) Handle(ctx context.Context, m Mail, fn func(context.Context, Mail) (Mail, error)) (Mail, error) {
	return mw(ctx, m, fn)
}

// Apply calls opt(d).
func (opt OptionFunc) Apply(d *Dog) {
	opt(d)
}

// Apply calls opt(d) for every Option opt in p.
func (p Plugin) Apply(d *Dog) {
	for _, opt := range p {
		opt.Apply(d)
	}
}

// Handle calls lis(ctx, h, m).
func (lis ListenerFunc) Handle(ctx context.Context, h Hook, m Mail) {
	lis(ctx, h, m)
}

const ctxSendError = ctxKey("sendError")

// SendError returns the error of the last (*Dog).Send() call that has been made using ctx.
func SendError(ctx context.Context) error {
	err, _ := ctx.Value(ctxSendError).(error)
	return err
}

func withSendError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ctxSendError, err)
}
