package office

//go:generate mockgen -source=office.go -destination=./mock_office/office.go

import (
	"context"
	"fmt"
	"sync"

	"github.com/bounoable/postdog/letter"
)

// ...
const (
	BeforeSend = SendHook(iota)
	AfterSend
)

var (
	defaultRunConfig = runConfig{
		workers: 1,
	}
)

// Office queues and dispatches outgoing letters.
// It is thread-safe (but the transports may not be).
type Office struct {
	cfg              Config
	mux              sync.RWMutex
	transports       map[string]Transport
	defaultTransport string
	queue            chan dispatchJob
}

// Config is the office configuration.
type Config struct {
	// QueueBuffer is the channel buffer size for outgoing letters.
	QueueBuffer int
	Middleware  []Middleware
	Logger      Logger
	SendHooks   map[SendHook][]func(context.Context, letter.Letter)
}

// Middleware ...
type Middleware interface {
	Handle(ctx context.Context, let letter.Letter) (letter.Letter, error)
}

// MiddlewareFunc ...
type MiddlewareFunc func(ctx context.Context, let letter.Letter) (letter.Letter, error)

// Handle ...
func (fn MiddlewareFunc) Handle(ctx context.Context, let letter.Letter) (letter.Letter, error) {
	return fn(ctx, let)
}

// Logger ...
type Logger interface {
	Log(v ...interface{})
}

// SendHook ...
type SendHook int

// Transport ...
type Transport interface {
	Send(context.Context, letter.Letter) error
}

type dispatchJob struct {
	letter    letter.Letter
	transport string
}

// New returns a new Office.
func New(opts ...Option) *Office {
	cfg := Config{
		Middleware: make([]Middleware, 0),
		SendHooks:  make(map[SendHook][]func(context.Context, letter.Letter)),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return NewWithConfig(cfg)
}

// NewWithConfig ...
func NewWithConfig(cfg Config) *Office {
	return &Office{
		cfg:        cfg,
		transports: make(map[string]Transport),
		queue:      make(chan dispatchJob, cfg.QueueBuffer),
	}
}

// Option ...
type Option func(*Config)

// QueueBuffer ...
func QueueBuffer(size int) Option {
	return func(cfg *Config) {
		cfg.QueueBuffer = size
	}
}

// WithMiddleware ...
func WithMiddleware(middleware ...Middleware) Option {
	return func(cfg *Config) {
		cfg.Middleware = append(cfg.Middleware, middleware...)
	}
}

// WithLogger ...
func WithLogger(logger Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

// WithHook ...
func WithHook(h SendHook, fns ...func(context.Context, letter.Letter)) Option {
	return func(cfg *Config) {
		cfg.SendHooks[h] = append(cfg.SendHooks[h], fns...)
	}
}

// Config returns the configration.
func (o *Office) Config() Config {
	return o.cfg
}

// ConfigureTransport configures a transport.
// The first transport will automatically be made the default transport,
// even if the Default() option is not used.
func (o *Office) ConfigureTransport(name string, trans Transport, options ...ConfigureOption) {
	var cfg configureConfig
	for _, opt := range options {
		opt(&cfg)
	}

	o.mux.Lock()
	defer o.mux.Unlock()

	if cfg.asDefault || len(o.transports) == 0 {
		o.defaultTransport = name
	}

	o.transports[name] = trans
}

type configureConfig struct {
	asDefault bool
}

// ConfigureOption ...
type ConfigureOption func(*configureConfig)

// DefaultTransport makes the transport the default transport.
func DefaultTransport() ConfigureOption {
	return func(cfg *configureConfig) {
		cfg.asDefault = true
	}
}

// Transport returns the configured transport for the given name.
func (o *Office) Transport(name string) (Transport, error) {
	o.mux.RLock()
	defer o.mux.RUnlock()
	trans, ok := o.transports[name]
	if !ok {
		return nil, UnconfiguredTransportError{
			Name: name,
		}
	}
	return trans, nil
}

// UnconfiguredTransportError ...
type UnconfiguredTransportError struct {
	Name string
}

func (err UnconfiguredTransportError) Error() string {
	return fmt.Sprintf("unconfigured transport: %s", err.Name)
}

// DefaultTransport returns the default transport.
func (o *Office) DefaultTransport() (Transport, error) {
	return o.Transport(o.defaultTransport)
}

// MakeDefault sets the default transport.
func (o *Office) MakeDefault(name string) error {
	if _, err := o.Transport(name); err != nil {
		return err
	}

	o.mux.Lock()
	defer o.mux.Unlock()
	o.defaultTransport = name

	return nil
}

// SendWith ...
func (o *Office) SendWith(ctx context.Context, transport string, let letter.Letter) error {
	trans, err := o.Transport(transport)
	if err != nil {
		return err
	}

	for _, mw := range o.cfg.Middleware {
		if let, err = mw.Handle(ctx, let); err != nil {
			return err
		}
	}

	for _, fn := range o.cfg.SendHooks[BeforeSend] {
		fn(ctx, let)
	}

	if err = trans.Send(ctx, let); err != nil {
		o.log(err)
	}

	for _, fn := range o.cfg.SendHooks[AfterSend] {
		fn(ctx, let)
	}

	return err
}

// Send ...
func (o *Office) Send(ctx context.Context, let letter.Letter) error {
	return o.SendWith(ctx, o.defaultTransport, let)
}

// Dispatch ...
func (o *Office) Dispatch(ctx context.Context, let letter.Letter, opts ...DispatchOption) error {
	job := dispatchJob{letter: let}
	for _, opt := range opts {
		opt(&job)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case o.queue <- job:
		return nil
	}
}

// DispatchOption ...
type DispatchOption func(*dispatchJob)

// DispatchWith ...
func DispatchWith(transport string) DispatchOption {
	return func(cfg *dispatchJob) {
		cfg.transport = transport
	}
}

// Run processes the outgoing letter queue.
func (o *Office) Run(ctx context.Context, opts ...RunOption) error {
	cfg := defaultRunConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var wg sync.WaitGroup
	wg.Add(cfg.workers)
	for i := 0; i < cfg.workers; i++ {
		go o.run(ctx, &wg)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (o *Office) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-o.queue:
			if job.transport == "" {
				o.Send(ctx, job.letter)
			} else {
				o.SendWith(ctx, job.transport, job.letter)
			}
		}
	}
}

func (o *Office) log(v ...interface{}) {
	if o.cfg.Logger == nil {
		return
	}
	o.cfg.Logger.Log(v...)
}

type runConfig struct {
	workers int
}

// RunOption ...
type RunOption func(*runConfig)

// Workers ...
func Workers(workers int) RunOption {
	return func(cfg *runConfig) {
		cfg.workers = workers
	}
}
