package store

//go:generate mockgen -source=store.go -destination=./mock_store/store.go

import (
	"context"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

// Store inserts letters into a repository.
type Store interface {
	Insert(context.Context, Letter) error
}

// Letter adds additional info to a letter.Letter.
type Letter struct {
	letter.Letter
	SentAt    time.Time
	SendError string
}

// Plugin is the install function for the store plugin.
// It hooks into the office.AfterSend hook and inserts all sent letters into the store implementation.
func Plugin(store Store) office.PluginFunc {
	return func(pctx office.PluginContext) {
		pctx.WithSendHook(office.AfterSend, func(ctx context.Context, let letter.Letter) {
			if Disabled(ctx) {
				return
			}

			var sendError string
			sendErr := office.SendError(ctx)
			if sendErr != nil {
				sendError = sendErr.Error()
			}

			if err := store.Insert(ctx, Letter{
				Letter:    let,
				SentAt:    time.Now(),
				SendError: sendError,
			}); err != nil {
				pctx.Log(err)
			}
		})
	}
}

// Disable disables the insertion of letters for ctx.
func Disable(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxDisabled, true)
}

// Disabled determines if the insertion of letters is disabled for ctx.
func Disabled(ctx context.Context) bool {
	disabled, _ := ctx.Value(ctxDisabled).(bool)
	return disabled
}

// Enable (re)enables the insertion of letters for ctx.
func Enable(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxDisabled, false)
}

type ctxKey string

var ctxDisabled = ctxKey("disabled")
