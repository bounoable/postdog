package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/office/mock_office"
	"github.com/bounoable/postdog/plugins/store"
	"github.com/bounoable/postdog/plugins/store/mock_store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
			repo.EXPECT().Insert(
				gomock.Any(),
				gomock.AssignableToTypeOf(store.Letter{}),
			).DoAndReturn(func(_ context.Context, slet store.Letter) error {
				assert.Equal(t, let, slet.Letter)
				assert.InDelta(t, time.Now().Unix(), slet.SentAt.Unix(), 1)
				return tcase.insertError
			})

			plugin := store.Plugin(repo)
			plugin(ctx)

			hookFn(context.Background(), let)
		})
	}
}
