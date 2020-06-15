package office

//go:generate mockgen -source=office.go -destination=./mock_office/office.go

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/bounoable/postdog/letter"
)

var (
	// DefaultLogger is the default logger implementation.
	// It redirects all Log(v ...interface{}) calls to log.Println(v...)
	DefaultLogger defaultLogger

	// NopLogger is a nil logger and effectively a no-op logger.
	NopLogger Logger

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

// Transport sends letters to the recipients.
type Transport interface {
	Send(context.Context, letter.Letter) error
}

type dispatchJob struct {
	letter    letter.Letter
	transport string
}

// New initializes a new *Office with opts.
func New(opts ...Option) *Office {
	cfg := Config{
		Middleware: make([]Middleware, 0),
		SendHooks:  make(map[SendHook][]func(context.Context, letter.Letter)),
		Logger:     DefaultLogger,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return NewWithConfig(cfg)
}

// NewWithConfig initializes a new *Office with the given cfg.
func NewWithConfig(cfg Config) *Office {
	off := &Office{
		cfg:        cfg,
		transports: make(map[string]Transport),
		queue:      make(chan dispatchJob, cfg.QueueBuffer),
	}

	ctx := pluginContext{cfg: &cfg}
	for _, plugin := range cfg.Plugins {
		plugin.Install(ctx)
	}

	return off
}

// Config returns the configration.
func (o *Office) Config() Config {
	return o.cfg
}

// ConfigureTransport configures the transport with the given name.
// The first configured transport is the default transport, even if the Default() option is not used.
// Subsequent calls to ConfigureTransport() with the Default() option override the default transport.
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

// ConfigureOption is a transport configuration option.
type ConfigureOption func(*configureConfig)

// DefaultTransport makes a transport the default transport.
func DefaultTransport() ConfigureOption {
	return func(cfg *configureConfig) {
		cfg.asDefault = true
	}
}

// Transport returns the configured transport for the given name.
// Returns an UnconfiguredTransportError, if the transport with the given name has not been registered.
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

// UnconfiguredTransportError means a transport has not been registered.
type UnconfiguredTransportError struct {
	Name string
}

func (err UnconfiguredTransportError) Error() string {
	return fmt.Sprintf("unconfigured transport: %s", err.Name)
}

// DefaultTransport returns the default transport.
// Returns an UnconfiguredTransportError if no transport has been configured.
func (o *Office) DefaultTransport() (Transport, error) {
	return o.Transport(o.defaultTransport)
}

// MakeDefault makes the transport with the given name the default transport.
func (o *Office) MakeDefault(name string) error {
	if _, err := o.Transport(name); err != nil {
		return err
	}

	o.mux.Lock()
	defer o.mux.Unlock()
	o.defaultTransport = name

	return nil
}

// SendWith sends a letter over the given transport.
// Returns an UnconfiguredTransportError, if the transport with the given name has not been registered.
// If a middleware returns an error, SendWith() will return that error.
// The BeforeSend hook is called after the middlewares, before Transport.Send().
// The AfterSend hook is called after Transport.Send(), even if Transport.Send() returns an error.
// Hooks are called concurrently.
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
		go fn(ctx, let)
	}

	if err = trans.Send(ctx, let); err != nil {
		o.log(err)
	}

	for _, fn := range o.cfg.SendHooks[AfterSend] {
		go fn(ctx, let)
	}

	return err
}

// Send calls SendWith() with the default transport.
func (o *Office) Send(ctx context.Context, let letter.Letter) error {
	return o.SendWith(ctx, o.defaultTransport, let)
}

// Dispatch adds let to the send queue with the given opts.
// Dispatch returns an error only if ctx is canceled before let has been queued.
// Available options:
//	DispatchWith(): Set the name of the transport to use.
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

// DispatchOption is a Dispatch() option.
type DispatchOption func(*dispatchJob)

// DispatchWith sets the transport to be used for sending the letter.
func DispatchWith(transport string) DispatchOption {
	return func(cfg *dispatchJob) {
		cfg.transport = transport
	}
}

// Run processes the outgoing letter queue with the gives options.
// Run blocks until ctx is canceled.
// Available options:
//	Workers(): Set the queue worker count.
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

// RunOption is a Run() option.
type RunOption func(*runConfig)

// Workers sets the worker count for the send queue.
func Workers(workers int) RunOption {
	return func(cfg *runConfig) {
		cfg.workers = workers
	}
}

type defaultLogger struct{}

func (l defaultLogger) Log(v ...interface{}) {
	log.Println(v...)
}
