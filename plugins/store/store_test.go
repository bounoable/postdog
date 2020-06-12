package store_test

import (
	"context"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/office/mock_office"
	"github.com/bounoable/postdog/plugins/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPlugin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := mock_office.NewMockPluginContext(ctrl)
	ctx.EXPECT().WithSendHook(office.AfterSend, gomock.Any()).DoAndReturn(func(
		h office.SendHook,
		_ func(context.Context, letter.Letter),
	) {
		assert.Equal(t, office.AfterSend, h)
	})

	store.Plugin(ctx)
}
