package nop_test

import (
	"context"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/transport/nop"
	"github.com/stretchr/testify/assert"
)

func TestNop_Send(t *testing.T) {
	ctx := context.Background()
	err := nop.Transport.Send(ctx, letter.Write())
	assert.Nil(t, err)
}
