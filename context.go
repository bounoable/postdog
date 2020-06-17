package postdog

import "context"

var (
	ctxSendError = ctxKey("send-error")
)

type ctxKey string

// SendError returns the send error from ctx.
// Returns nil, if ctx contains no send error.
func SendError(ctx context.Context) error {
	err, _ := ctx.Value(ctxSendError).(error)
	return err
}
