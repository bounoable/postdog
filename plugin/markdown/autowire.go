package markdown

import (
	"context"
	"fmt"
	"sync"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
)

var (
	// Name is the plugin name.
	Name = "markdown"
)

var (
	convertersMux sync.RWMutex
	converters    = map[string]ConverterFactory{}
)

// ConverterFactory creates the Markdown converter from the autowire config.
type ConverterFactory interface {
	CreateConverter(cfg map[string]interface{}) (Converter, error)
}

// ConverterFactoryFunc allows functions to be used as a ConverterFactory.
type ConverterFactoryFunc func(map[string]interface{}) (Converter, error)

// CreateConverter creates the Markdown converter from the autowire config.
func (fn ConverterFactoryFunc) CreateConverter(cfg map[string]interface{}) (Converter, error) {
	return fn(cfg)
}

// RegisterConverter registers a converter factory for the given converter name for autowiring.
func RegisterConverter(name string, factory ConverterFactory) {
	convertersMux.Lock()
	defer convertersMux.Unlock()
	converters[name] = factory
}

// Register registers the plugin factory in the autowire config.
func Register(cfg *autowire.Config) {
	cfg.RegisterPlugin(Name, autowire.PluginFactoryFunc(AutowirePlugin))
}

// AutowirePlugin creates the Markdown plugin from the autowire config.
func AutowirePlugin(_ context.Context, cfg map[string]interface{}) (postdog.Plugin, error) {
	converterName, ok := cfg["use"].(string)
	overrideHTML, _ := cfg["overrideHTML"].(bool)

	convertersMux.RLock()
	defer convertersMux.RUnlock()

	factory, ok := converters[converterName]
	if !ok {
		return nil, UnregisteredConverterError{
			Name: converterName,
		}
	}

	conv, err := factory.CreateConverter(cfg)
	if err != nil {
		return nil, err
	}

	return Plugin(conv, OverrideHTML(overrideHTML)), nil
}

// UnregisteredConverterError means the autowire config defines the plugin with an unregistered converter name.
type UnregisteredConverterError struct {
	Name string
}

func (err UnregisteredConverterError) Error() string {
	return fmt.Sprintf("unregistered converter: %s", err.Name)
}
