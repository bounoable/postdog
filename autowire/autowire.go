// Package autowire provides office initialization through a YAML config.
package autowire

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/bounoable/postdog/office"
	"gopkg.in/yaml.v3"
)

// Load reads the configuration from r and returns the autowire config.
func Load(r io.Reader, opts ...Option) (Config, error) {
	cfg := New(opts...)
	return cfg, cfg.Load(r)
}

// File reads the configuration from the file at path and returns the autowire config.
func File(path string, opts ...Option) (Config, error) {
	cfg := New(opts...)
	return cfg, cfg.LoadFile(path)
}

// Config is the autowire configuration.
// You should use the Load() or File() functions to build the configuration.
type Config struct {
	Providers  map[string]TransportFactory
	Transports map[string]TransportConfig
}

// TransportFactory creates transports from user-provided configuration.
type TransportFactory interface {
	CreateTransport(ctx context.Context, cfg map[string]interface{}) (office.Transport, error)
}

// The TransportFactoryFunc allows a transport factory function to be used as a TransportFactory.
type TransportFactoryFunc func(context.Context, map[string]interface{}) (office.Transport, error)

// CreateTransport creates a transport from user-provided configuration.
func (fn TransportFactoryFunc) CreateTransport(ctx context.Context, cfg map[string]interface{}) (office.Transport, error) {
	return fn(ctx, cfg)
}

// TransportConfig contains the parsed used-provided configuration for a single transport.
type TransportConfig struct {
	Provider string
	Config   map[string]interface{}
}

// New initializes a new autowire configuration.
// Instead of calling New() directly, you should use Load() or File() instead.
func New(opts ...Option) Config {
	cfg := Config{
		Providers:  make(map[string]TransportFactory),
		Transports: make(map[string]TransportConfig),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

// Option is an autowire constructor option.
type Option func(*Config)

// Provider registers a transport factory for the given provider name.
// Providers have to be registered in order to be used in the configuration file.
func Provider(name string, factory TransportFactory) Option {
	return func(cfg *Config) {
		cfg.Providers[name] = factory
	}
}

// RegisterProvider registers a transport factory for the given provider name.
func (cfg Config) RegisterProvider(name string, factory TransportFactory) {
	cfg.Providers[name] = factory
}

// Get returns the parsed configuration for the given transport name.
// Calling Get() with an unconfigured transport name results in an UnconfiguredTransportError.
func (cfg Config) Get(name string) (TransportConfig, error) {
	tcfg, ok := cfg.Transports[name]
	if !ok {
		return TransportConfig{}, UnconfiguredTransportError{
			Name: name,
		}
	}
	return tcfg, nil
}

// UnconfiguredTransportError is returned by Config.Get() if the given transport name hasn't been configured yet.
type UnconfiguredTransportError struct {
	Name string
}

func (err UnconfiguredTransportError) Error() string {
	return fmt.Sprintf("unconfigured transport: %s", err.Name)
}

// LoadFile loads the YAML autowire configuration from the file at the given path.
func (cfg Config) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return cfg.Load(f)
}

// Load loads the YAML autowire configuration from r.
func (cfg Config) Load(r io.Reader) error {
	var yamlCfg yamlConfig
	if err := yaml.NewDecoder(r).Decode(&yamlCfg); err != nil {
		return err
	}
	return yamlCfg.apply(cfg)
}

type yamlConfig struct {
	// map[TRANSPORT_NAME]map[CONFIGKEY]interface{}
	Transports map[string]map[string]interface{}
	// Default is the name of the default transport.
	Default string
}

func (cfg yamlConfig) apply(config Config) error {
	transports := make(map[string]TransportConfig)

	for name, transportcfg := range cfg.Transports {
		if _, ok := transports[name]; ok {
			return DuplicateTransportError{Name: name}
		}

		provider, ok := transportcfg["provider"].(string)
		if !ok {
			return InvalidConfigError{
				Transport: name,
				ConfigKey: "provider",
				Expected:  "",
				Provided:  provider,
			}
		}

		varcfg := make(map[string]interface{})

		if ivarcfg, ok := transportcfg["config"]; ok {
			tcfg, ok := ivarcfg.(map[string]interface{})
			if !ok {
				return InvalidConfigError{
					Transport: name,
					ConfigKey: "config",
					Expected:  new(map[string]interface{}),
					Provided:  tcfg,
				}
			}
			varcfg = tcfg
		}

		applyEnvVars(varcfg)

		transports[name] = TransportConfig{
			Provider: provider,
			Config:   varcfg,
		}
	}

	for name, transportcfg := range transports {
		config.Transports[name] = transportcfg
	}

	return nil
}

// DuplicateTransportError means the YAML configuration contains multiple configurations for the same transport name.
type DuplicateTransportError struct {
	Name string
}

func (err DuplicateTransportError) Error() string {
	return fmt.Sprintf("duplicate transport name: %s", err.Name)
}

// InvalidConfigError means the configuration for a transport contains an invalid value.
type InvalidConfigError struct {
	Transport string
	ConfigKey string
	Expected  interface{}
	Provided  interface{}
}

func (err InvalidConfigError) Error() string {
	return fmt.Sprintf("invalid config value for carrier '%s': '%s' must be a '%T' but is a '%T'", err.Transport, err.ConfigKey, err.Expected, err.Provided)
}

func applyEnvVars(cfg map[string]interface{}) {
	for key, val := range cfg {
		switch v := val.(type) {
		case map[string]interface{}:
			applyEnvVars(v)
		case string:
			cfg[key] = replaceEnvPlaceholders(v)
		}
	}
}

var envPlaceholderExpr = regexp.MustCompile(`(?Ui)\${(.+)}`)

func replaceEnvPlaceholders(val string) string {
	return envPlaceholderExpr.ReplaceAllStringFunc(val, func(placeholder string) string {
		return os.Getenv(envPlaceholderExpr.ReplaceAllString(placeholder, "$1"))
	})
}

// Office builds the *office.Office from the autowire configuration.
// You have to register the used providers with the provided opts.
func (cfg Config) Office(ctx context.Context, opts ...office.Option) (*office.Office, error) {
	off := office.New(opts...)
	for name, transportcfg := range cfg.Transports {
		factory, ok := cfg.Providers[transportcfg.Provider]
		if !ok {
			return nil, UnregisteredProviderError{
				Name: transportcfg.Provider,
			}
		}

		trans, err := factory.CreateTransport(ctx, transportcfg.Config)
		if err != nil {
			return nil, fmt.Errorf("could not create transport '%s': %w", name, err)
		}

		off.ConfigureTransport(name, trans)
	}
	return off, nil
}

// UnregisteredProviderError means the autowire configuration uses a provider that hasn't been registered.
type UnregisteredProviderError struct {
	Name string
}

func (err UnregisteredProviderError) Error() string {
	return fmt.Sprintf("unregistered provider: %s", err.Name)
}
