// Package autowire provides office initialization through a YAML config.
package autowire

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/bounoable/postdog"
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
	TransportFactories map[string]TransportFactory
	Transports         map[string]TransportConfig
	DefaultTransport   string
	PluginFactories    map[string]PluginFactory
	Plugins            []PluginConfig
	Queue              QueueConfig
}

// TransportFactory create the transport from the given cfg.
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

// PluginFactory creates the plugin from the given cfg.
type PluginFactory interface {
	CreatePlugin(ctx context.Context, cfg map[string]interface{}) (postdog.Plugin, error)
}

// PluginFactoryFunc allows functions to be used as a PluginFactory.
type PluginFactoryFunc func(context.Context, map[string]interface{}) (postdog.Plugin, error)

// CreatePlugin creates the plugin from the given cfg.
func (fn PluginFactoryFunc) CreatePlugin(ctx context.Context, cfg map[string]interface{}) (postdog.Plugin, error) {
	return fn(ctx, cfg)
}

// PluginConfig is the configuration for a plugin.
type PluginConfig struct {
	Name   string
	Config map[string]interface{}
}

// QueueConfig is the send queue configuration.
type QueueConfig struct {
	Buffer int
}

// New initializes a new autowire configuration.
// Instead of calling New() directly, you should use Load() or File() instead.
func New(opts ...Option) Config {
	cfg := Config{
		TransportFactories: make(map[string]TransportFactory),
		Transports:         make(map[string]TransportConfig),
		PluginFactories:    make(map[string]PluginFactory),
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
		cfg.TransportFactories[name] = factory
	}
}

// RegisterProvider registers a transport factory for the given provider name.
func (cfg Config) RegisterProvider(name string, factory TransportFactory) {
	cfg.TransportFactories[name] = factory
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

// RegisterPlugin registers the factory for the given plugin name.
func (cfg Config) RegisterPlugin(name string, factory PluginFactory) {
	cfg.PluginFactories[name] = factory
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
	Queue   QueueConfig
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

	config.DefaultTransport = replaceEnvPlaceholders(cfg.Default)

	for _, plugincfg := range cfg.Plugins {
		pcfg := make(map[string]interface{}, len(plugincfg.Config))
		for key, val := range plugincfg.Config {
			pcfg[key] = val
		}
		applyEnvVars(pcfg)
		plugincfg.Config = pcfg
		config.Plugins = append(config.Plugins, plugincfg)
	}

	config.Queue = cfg.Queue

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
// You have to register the used providers with opts, if they haven't been globally registered.
func (cfg Config) Office(ctx context.Context, opts ...postdog.Option) (*postdog.Office, error) {
	var pluginOpts []postdog.Option

	for _, plugincfg := range cfg.Plugins {
		factory, ok := cfg.PluginFactories[plugincfg.Name]
		if !ok {
			return nil, UnregisteredPluginError{
				Name: plugincfg.Name,
			}
		}

		plugin, err := factory.CreatePlugin(ctx, plugincfg.Config)
		if err != nil {
			return nil, fmt.Errorf("could not create plugin '%s': %w", plugincfg.Name, err)
		}

		pluginOpts = append(pluginOpts, postdog.WithPlugin(plugin))
	}

	opts = append([]postdog.Option{
		postdog.QueueBuffer(cfg.Queue.Buffer),
	}, opts...)

	off := postdog.New(append(pluginOpts, opts...)...)

	for name, transportcfg := range cfg.Transports {
		factory, ok := cfg.TransportFactories[transportcfg.Provider]
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

// UnregisteredPluginError means the autowire configuration uses a plugin that hasn't been registered.
type UnregisteredPluginError struct {
	Name string
}

func (err UnregisteredPluginError) Error() string {
	return fmt.Sprintf("unregistered plugin: %s", err.Name)
}
