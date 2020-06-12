package store

//go:generate mockgen -source=store.go -destination=./mock_store/store.go

import (
	"context"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

// Store ...
type Store interface {
	Insert(context.Context, letter.Letter) error
}

// Plugin ...
func Plugin(store Store) office.PluginFunc {
	return func(pctx office.PluginContext) {
		pctx.WithSendHook(office.AfterSend, func(ctx context.Context, let letter.Letter) {
			if err := store.Insert(ctx, let); err != nil {
				pctx.Log(err)
			}
		})
	}
}
