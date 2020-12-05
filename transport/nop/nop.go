package nop

import (
	"context"

	"github.com/bounoable/postdog"
)

// Transport is a no-op transport.
var Transport transport

type transport struct{}

func (tr transport) Send(context.Context, postdog.Mail) error {
	return nil
}

// Factory just returns Transport.
func Factory(context.Context, map[string]interface{}) (postdog.Transport, error) {
	return Transport, nil
}
