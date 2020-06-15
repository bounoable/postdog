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
	SentAt time.Time
}

// Plugin is the install function for the store plugin.
// It hooks into the office.AfterSend hook and inserts all sent letters into the store implementation.
func Plugin(store Store) office.PluginFunc {
	return func(pctx office.PluginContext) {
		// TODO: Send error handling
		pctx.WithSendHook(office.AfterSend, func(ctx context.Context, let letter.Letter) {
			if err := store.Insert(ctx, Letter{
				Letter: let,
				SentAt: time.Now(),
			}); err != nil {
				pctx.Log(err)
			}
		})
	}
}
