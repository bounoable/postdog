package transport_test

import (
	"context"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/transport"
	"github.com/stretchr/testify/assert"
)

func TestNop_Send(t *testing.T) {
	ctx := context.Background()
	err := transport.Nop.Send(ctx, letter.Write())
	assert.Nil(t, err)
}
