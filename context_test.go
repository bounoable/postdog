package postdog

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestSendTime(t *testing.T) {
	ctx := context.Background()
	st := SendTime(ctx)
	assert.True(t, st.IsZero())
	want := time.Now()
	ctx = withSendTime(ctx, want)
	st = SendTime(ctx)
	assert.Equal(t, want, st)
}
