package store_test

import (
	"context"
	"errors"
	"sync"
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
	sendErrorStub   = errors.New("send error")
	insertErrorStub = errors.New("insert error")
)

func TestPlugin(t *testing.T) {
	cases := map[string]struct {
		sendError   error
		insertError error
	}{
		"works": {},
		"send error": {
			sendError: sendErrorStub,
		},
		"insert error": {
			insertError: insertErrorStub,
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var wg sync.WaitGroup

			logger := mock_office.NewMockLogger(ctrl)
			let := letter.Write(letter.Subject("Hello"))

			if tcase.sendError != nil {
				logger.EXPECT().Log(tcase.sendError)
			}

			if tcase.insertError != nil {
				logger.EXPECT().Log(tcase.insertError)
			}

			wg.Add(1)
			repo := mock_store.NewMockStore(ctrl)
			repo.EXPECT().Insert(
				gomock.Any(),
				gomock.AssignableToTypeOf(store.Letter{}),
			).DoAndReturn(func(_ context.Context, slet store.Letter) error {
				defer wg.Done()

				assert.Equal(t, let, slet.Letter)
				assert.InDelta(t, time.Now().Unix(), slet.SentAt.Unix(), 1)

				var expectedErr string
				if tcase.sendError != nil {
					expectedErr = tcase.sendError.Error()
				}
				assert.Equal(t, expectedErr, slet.SendError)

				return tcase.insertError
			})

			trans := mock_office.NewMockTransport(ctrl)
			trans.EXPECT().Send(context.Background(), let).Return(tcase.sendError)

			off := office.New(
				office.WithLogger(logger),
				office.WithPlugin(store.Plugin(repo)),
			)
			off.ConfigureTransport("test", trans)

			err := off.Send(context.Background(), let)
			assert.Equal(t, tcase.sendError, err)

			wg.Wait()
		})
	}
}

func TestDisable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	assert.False(t, store.Disabled(ctx))
	ctx = store.Disable(ctx)
	assert.True(t, store.Disabled(ctx))
	ctx = store.Enable(ctx)
	assert.False(t, store.Disabled(ctx))
	ctx = store.Disable(ctx)

	repo := mock_store.NewMockStore(ctrl)
	off := office.New(office.WithPlugin(store.Plugin(repo)))
	trans := mock_office.NewMockTransport(ctrl)
	trans.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	off.ConfigureTransport("test", trans)

	off.Send(ctx, letter.Write())
}
