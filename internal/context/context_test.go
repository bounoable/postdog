package context_test

import (
	stdcontext "context"
	"errors"
	"testing"
	"time"

	"github.com/bounoable/postdog/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestSendError(t *testing.T) {
	ctx := stdcontext.Background()
	err := context.SendError(ctx)
	assert.Nil(t, err)

	want := errors.New("send error")
	ctx = context.WithSendError(ctx, want)
	err = context.SendError(ctx)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, want))
}

func TestSendTime(t *testing.T) {
	ctx := stdcontext.Background()
	st := context.SendTime(ctx)
	assert.True(t, st.IsZero())
	want := time.Now()
	ctx = context.WithSendTime(ctx, want)
	st = context.SendTime(ctx)
	assert.Equal(t, want, st)
}
