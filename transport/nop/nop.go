package nop

import (
	"context"

	"github.com/bounoable/postdog/letter"
)

var (
	// Provider is the provider name for the no-op transport.
	Provider = "nop"

	// Transport is the no-op transport.
	Transport = transport{}
)

type transport struct{}

func (trans transport) Send(_ context.Context, _ letter.Letter) error {
	return nil
}
