package postdog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendError(t *testing.T) {
	ctx := context.Background()
	err := SendError(ctx)
	assert.Nil(t, err)

	want := errors.New("send error")
	ctx = withSendError(ctx, want)
	err = SendError(ctx)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, want))
}
