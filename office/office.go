package office

import (
	"context"
	"fmt"
	"sync"

	"github.com/bounoable/postdog/letter"
)

//go:generate mockgen -source=office.go -destination=./mock_office/office.go

// Office queues and dispatches outgoing letters.
// It is thread-safe (but the transports may not be).
type Office struct {
	mux              sync.RWMutex
	transports       map[string]Transport
	defaultTransport string
}

// Transport ...
type Transport interface {
	Send(context.Context, *letter.Letter) error
}

// New returns a new Office.
func New() *Office {
	return &Office{
		transports: make(map[string]Transport),
	}
}

// Configure configures a transport.
// The first transport will automatically be made the default transport,
// even if the Default() option is not used.
func (o *Office) Configure(name string, trans Transport, options ...ConfigureOption) {
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
