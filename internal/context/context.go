package context

import (
	"context"
	"time"
)

type ctxKey string

const ctxSendError = ctxKey("sendError")

// SendError returns the error of the last (*Dog).Send() call that has been made using ctx.
func SendError(ctx context.Context) error {
	err, _ := ctx.Value(ctxSendError).(error)
	return err
}

// WithSendError returns a copy of ctx with the send error set to err.
func WithSendError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ctxSendError, err)
}

const ctxSendTime = ctxKey("sendTime")

// SendTime returns the time of the last (*Dog).Send() call that has been made using ctx.
func SendTime(ctx context.Context) time.Time {
	t, _ := ctx.Value(ctxSendTime).(time.Time)
	return t
}

// WithSendTime returns a copy of ctx with the send time set to t.
func WithSendTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, ctxSendTime, t)
}
