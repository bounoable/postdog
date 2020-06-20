package transport

import (
	"context"

	"github.com/bounoable/postdog/letter"
)

// Nop is the no-op transport.
var Nop = nopTransport{}

type nopTransport struct{}

func (trans nopTransport) Send(_ context.Context, _ letter.Letter) error {
	return nil
}
