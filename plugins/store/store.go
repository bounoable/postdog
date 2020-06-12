package store

import (
	"context"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

// Plugin ...
func Plugin(ctx office.PluginContext) {
	ctx.WithSendHook(office.AfterSend, func(_ context.Context, _ letter.Letter) {

	})
}
