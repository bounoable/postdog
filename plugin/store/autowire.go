package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
)

var (
	// Name is the plugin name.
	Name = "store"

	providersMux sync.RWMutex
	providers    = map[string]Factory{}
)

func init() {
	autowire.RegisterPlugin(Name, autowire.PluginFactoryFunc(AutowirePlugin))
}

// Factory is a store factory.
type Factory interface {
	CreateStore(context.Context, map[string]interface{}) (Store, error)
}

// FactoryFunc allows functions to be used as a Factory.
type FactoryFunc func(context.Context, map[string]interface{}) (Store, error)

// CreateStore creates the store from the given cfg.
func (fn FactoryFunc) CreateStore(ctx context.Context, cfg map[string]interface{}) (Store, error) {
	return fn(ctx, cfg)
}

// RegisterProvider globally registers the store factory for the given store provider name for autowiring.
func RegisterProvider(name string, factory Factory) {
	providersMux.Lock()
	defer providersMux.Unlock()
	providers[name] = factory
}

// Register registers the plugin in the autowire config.
func Register(cfg *autowire.Config) {
	cfg.RegisterPlugin(Name, autowire.PluginFactoryFunc(AutowirePlugin))
}

// AutowirePlugin creates the store plugin from the given cfg.
func AutowirePlugin(ctx context.Context, cfg map[string]interface{}) (postdog.Plugin, error) {
	providerName, _ := cfg["use"].(string)

	providersMux.RLock()
	defer providersMux.RUnlock()

	factory, ok := providers[providerName]
	if !ok {
		return nil, UnregisteredProviderError{
			Name: providerName,
		}
	}

	store, err := factory.CreateStore(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create store '%s': %w", providerName, err)
	}

	return Plugin(store), nil
}

// UnregisteredProviderError means the autowire config defines the plugin with an unregistered store provider name.
type UnregisteredProviderError struct {
	Name string
}

func (err UnregisteredProviderError) Error() string {
	return fmt.Sprintf("unregistered provider: %s", err.Name)
}
