package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/office/mock_office"
	"github.com/bounoable/postdog/plugins/store"
	"github.com/bounoable/postdog/plugins/store/mock_store"
	"github.com/golang/mock/gomock"
)

var (
	insertErrorStub = errors.New("insert error")
)

func TestPlugin(t *testing.T) {
	cases := map[string]struct {
		insertError error
	}{
		"works": {},
		"insert error": {
			insertError: insertErrorStub,
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			let := letter.Write(letter.Subject("Hello"))

			var hookFn func(context.Context, letter.Letter)

			ctx := mock_office.NewMockPluginContext(ctrl)
			ctx.EXPECT().WithSendHook(office.AfterSend, gomock.Any()).DoAndReturn(func(
				_ office.SendHook,
				fn func(ctx context.Context, let letter.Letter),
			) {
				hookFn = fn
			})

			if tcase.insertError != nil {
				ctx.EXPECT().Log(tcase.insertError)
			}

			repo := mock_store.NewMockStore(ctrl)
			repo.EXPECT().Insert(gomock.Any(), let).Return(tcase.insertError)

			plugin := store.Plugin(repo)
			plugin(ctx)

			hookFn(context.Background(), let)
		})
	}
}
