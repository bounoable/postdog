// Package autowire provides office initialization through a YAML config.
package autowire

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"

	"github.com/bounoable/postdog"
	"gopkg.in/yaml.v3"
)

var (
	globalProvidersMux sync.RWMutex
	globalProviders    = map[string]TransportFactory{}
)

// RegisterProvider globally registers a transport factory for the given provider name.
func RegisterProvider(name string, factory TransportFactory) {
	globalProvidersMux.Lock()
	defer globalProvidersMux.Unlock()
	globalProviders[name] = factory
}

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
	Plugins    []PluginConfig
}

// TransportFactory creates transports from user-provided configuration.
type TransportFactory interface {
	CreateTransport(ctx context.Context, cfg map[string]interface{}) (postdog.Transport, error)
}

// The TransportFactoryFunc allows a transport factory function to be used as a TransportFactory.
type TransportFactoryFunc func(context.Context, map[string]interface{}) (postdog.Transport, error)

// CreateTransport creates a transport from user-provided configuration.
func (fn TransportFactoryFunc) CreateTransport(ctx context.Context, cfg map[string]interface{}) (postdog.Transport, error) {
	return fn(ctx, cfg)
}

// TransportConfig contains the parsed used-provided configuration for a single transport.
type TransportConfig struct {
	Provider string
	Config   map[string]interface{}
}

// PluginConfig is the configuration for a plugin.
type PluginConfig struct {
	Name   string
	Config map[string]interface{}
}

// New initializes a new autowire configuration.
// Instead of calling New() directly, you should use Load() or File() instead.
func New(opts ...Option) Config {
	cfg := Config{
		Providers:  make(map[string]TransportFactory),
		Transports: make(map[string]TransportConfig),
	}

	globalProvidersMux.RLock()
	defer globalProvidersMux.RUnlock()

	for name, factory := range globalProviders {
		cfg.Providers[name] = factory
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
func (cfg *Config) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return cfg.Load(f)
}

// Load loads the YAML autowire configuration from r.
func (cfg *Config) Load(r io.Reader) error {
	var yamlCfg yamlConfig
	if err := yaml.NewDecoder(r).Decode(&yamlCfg); err != nil {
		return err
	}
	return yamlCfg.apply(cfg)
}

type yamlConfig struct {
	// map[TRANSPORT_NAME]map[CONFIGKEY]interface{}
	Transports map[string]map[string]interface{}
	Plugins    []PluginConfig
	// Default is the name of the default transport.
	Default string
}

func (cfg yamlConfig) apply(config *Config) error {
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

	config.Plugins = append(config.Plugins, cfg.Plugins...)

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

// Office builds the *postdog.Office from the autowire configuration.
// You have to register the used providers with the provided opts.
func (cfg Config) Office(ctx context.Context, opts ...postdog.Option) (*postdog.Office, error) {
	off := postdog.New(opts...)
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
