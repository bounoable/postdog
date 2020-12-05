package config

//go:generate mockgen -source=config.go -destination=./mocks/config.go

import (
	"context"
	"errors"
	"fmt"

	"github.com/bounoable/postdog"
	"gopkg.in/yaml.v3"
)

var (
	// ErrUnknownTransport means a TransportFactory is missing for a transport.
	ErrUnknownTransport = errors.New("unknown transport")
)

// Config is the postdog configuration.
type Config struct {
	transports         map[string]Transport
	transportFactories map[string]TransportFactory
}

// Transport is a transport configuration.
type Transport struct {
	Use    string                 `yaml:"use"`
	Config map[string]interface{} `yaml:"config"`
}

// A TransportFactory accepts the transport-specific configuration and instantiates a transport from that configuration.
type TransportFactory interface {
	Transport(context.Context, map[string]interface{}) (postdog.Transport, error)
}

// Parse parses the YAML configuration in raw.
func (cfg *Config) Parse(raw []byte) error {
	var rawCfg rawConfig
	if err := yaml.Unmarshal(raw, &rawCfg); err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}
	cfg.transports = rawCfg.Transports
	return nil
}

// Transport returns the transport configuration for the given name,
// or ok=false if the config doesn't have a transport with that name.
func (cfg *Config) Transport(name string) (tr Transport, ok bool) {
	if cfg.transports == nil {
		return
	}
	tr, ok = cfg.transports[name]
	return
}

// Dog instantiates the *postdog.Dog from the parsed configuration.
//
// For every distinct `transport.use` config value a TransportFactory must be
// provided. It will return ErrUnknownTransport if a TransportFactory is missing.
func (cfg *Config) Dog(ctx context.Context, opts ...Option) (*postdog.Dog, error) {
	var dogOpts []postdog.Option

	cfg.transportFactories = make(map[string]TransportFactory)
	for _, opt := range opts {
		opt(cfg)
	}

	for name, transportConfig := range cfg.transports {
		factory, ok := cfg.transportFactories[transportConfig.Use]
		if !ok {
			return nil, ErrUnknownTransport
		}
		factoryConfig := transportConfig.Config
		if factoryConfig == nil {
			factoryConfig = make(map[string]interface{})
		}
		tr, err := factory.Transport(ctx, factoryConfig)
		if err != nil {
			return nil, fmt.Errorf("make transport %s: %w", name, err)
		}
		dogOpts = append(dogOpts, postdog.WithTransport(name, tr))
	}

	return postdog.New(dogOpts...), nil
}

// Option is an option for the (*Config).Dog() method.
type Option func(*Config)

// WithTransportFactory returns an Option that specifies the TransportFactory for a `transport.use` value.
func WithTransportFactory(use string, factory TransportFactory) Option {
	return func(cfg *Config) {
		cfg.transportFactories[use] = factory
	}
}

type rawConfig struct {
	Transports map[string]Transport `yaml:"transports"`
}
