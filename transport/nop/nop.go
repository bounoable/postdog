package nop

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
	"github.com/bounoable/postdog/letter"
)

var (
	// Provider is the provider name for the no-op transport.
	Provider = "nop"

	// Transport is the no-op transport.
	Transport = transport{}
)

// Register registers the no-op transpot in the autowire config.
func Register(cfg *autowire.Config) {
	cfg.RegisterProvider(Provider, autowire.TransportFactoryFunc(
		func(_ context.Context, _ map[string]interface{}) (postdog.Transport, error) {
			return Transport, nil
		}),
	)
}

type transport struct{}

func (trans transport) Send(_ context.Context, _ letter.Letter) error {
	return nil
}
