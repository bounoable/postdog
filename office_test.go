package postdog_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/mock_postdog"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	off := postdog.New(postdog.QueueBuffer(12))

	assert.Equal(t, postdog.Config{
		QueueBuffer: 12,
		Middleware:  make([]postdog.Middleware, 0),
		SendHooks:   make(map[postdog.SendHook][]func(context.Context, letter.Letter)),
		Logger:      postdog.DefaultLogger,
	}, off.Config())
}

func TestNew_withPlugin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	plugin := mock_postdog.NewMockPlugin(ctrl)
	plugin.EXPECT().Install(gomock.Any())

	postdog.New(postdog.WithPlugin(plugin))
}

func TestOffice_ConfigureTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	off := postdog.New()
	_, err := off.Transport("test")

	assert.True(t, errors.Is(err, postdog.UnconfiguredTransportError{
		Name: "test",
	}))

	_, err = off.DefaultTransport()

	assert.True(t, errors.Is(err, postdog.UnconfiguredTransportError{}))

	mockTrans := mock_postdog.NewMockTransport(ctrl)
	off.ConfigureTransport("test", mockTrans)
	trans, err := off.Transport("test")

	assert.Nil(t, err)
	assert.Equal(t, mockTrans, trans)

	defaultTrans, err := off.DefaultTransport()

	assert.Nil(t, err)
	assert.Equal(t, mockTrans, defaultTrans)
}

func TestOffice_ConfigureTransport_asDefault(t *testing.T) {
	cases := map[string]struct {
		configure func(*postdog.Office, *gomock.Controller)
		expected  string
	}{
		"default default": {
			configure: func(off *postdog.Office, ctrl *gomock.Controller) {
				off.ConfigureTransport("test1", mock_postdog.NewMockTransport(ctrl))
				off.ConfigureTransport("test2", mock_postdog.NewMockTransport(ctrl))
				off.ConfigureTransport("test3", mock_postdog.NewMockTransport(ctrl))
			},
			expected: "test1",
		},
		"first as default": {
			configure: func(off *postdog.Office, ctrl *gomock.Controller) {
				off.ConfigureTransport("test1", mock_postdog.NewMockTransport(ctrl), postdog.DefaultTransport())
				off.ConfigureTransport("test2", mock_postdog.NewMockTransport(ctrl))
				off.ConfigureTransport("test3", mock_postdog.NewMockTransport(ctrl))
			},
			expected: "test1",
		},
		"other-than-first as default": {
			configure: func(off *postdog.Office, ctrl *gomock.Controller) {
				off.ConfigureTransport("test1", mock_postdog.NewMockTransport(ctrl))
				off.ConfigureTransport("test2", mock_postdog.NewMockTransport(ctrl), postdog.DefaultTransport())
				off.ConfigureTransport("test3", mock_postdog.NewMockTransport(ctrl))
			},
			expected: "test2",
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			off := postdog.New()
			tcase.configure(off, ctrl)

			expected, err := off.Transport(tcase.expected)
			assert.Nil(t, err)
			trans, err := off.DefaultTransport()
			assert.Nil(t, err)

			assert.Equal(t, expected, trans)
		})
	}
}

func TestOffice_MakeDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	off := postdog.New()
	off.ConfigureTransport("test1", mock_postdog.NewMockTransport(ctrl))
	off.ConfigureTransport("test2", mock_postdog.NewMockTransport(ctrl))
	off.ConfigureTransport("test3", mock_postdog.NewMockTransport(ctrl))

	assertDefaultTransport(t, off, "test1")

	err := off.MakeDefault("test1")
	assert.Nil(t, err)
	assertDefaultTransport(t, off, "test1")

	err = off.MakeDefault("test2")
	assert.Nil(t, err)
	assertDefaultTransport(t, off, "test2")

	err = off.MakeDefault("test3")
	assert.Nil(t, err)
	assertDefaultTransport(t, off, "test3")

	err = off.MakeDefault("test4")
	assert.True(t, errors.Is(err, postdog.UnconfiguredTransportError{Name: "test4"}))
}

func assertDefaultTransport(t *testing.T, off *postdog.Office, name string) {
	trans, err := off.DefaultTransport()
	assert.Nil(t, err)

	expected, err := off.Transport(name)
	assert.Nil(t, err)

	assert.Equal(t, expected, trans)
}

func TestOffice_SendWith(t *testing.T) {
	cases := map[string]struct {
		configure   func(*postdog.Office, *gomock.Controller)
		name        string
		expectedErr error
	}{
		"unconfigured transport": {
			name: "test",
			expectedErr: postdog.UnconfiguredTransportError{
				Name: "test",
			},
		},
		"configured transport": {
			configure: func(off *postdog.Office, ctrl *gomock.Controller) {
				trans := mock_postdog.NewMockTransport(ctrl)
				trans.EXPECT().Send(
					gomock.AssignableToTypeOf(context.Background()),
					gomock.AssignableToTypeOf(letter.Write()),
				)
				off.ConfigureTransport("test", trans)
			},
			name: "test",
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			off := postdog.New()
			if tcase.configure != nil {
				tcase.configure(off, ctrl)
			}

			err := off.SendWith(context.Background(), tcase.name, letter.Write())

			assert.True(t, errors.Is(err, tcase.expectedErr))
		})
	}
}

func TestOffice_Send(t *testing.T) {
	cases := map[string]struct {
		configure   func(*postdog.Office, *gomock.Controller)
		expectedErr error
	}{
		"no transport configured": {
			expectedErr: postdog.UnconfiguredTransportError{},
		},
		"default transport": {
			configure: func(off *postdog.Office, ctrl *gomock.Controller) {
				trans := mock_postdog.NewMockTransport(ctrl)
				trans.EXPECT().Send(
					gomock.AssignableToTypeOf(context.Background()),
					gomock.AssignableToTypeOf(letter.Write()),
				)
				off.ConfigureTransport("test", trans)
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			off := postdog.New()
			if tcase.configure != nil {
				tcase.configure(off, ctrl)
			}

			err := off.Send(context.Background(), letter.Write())
			assert.True(t, errors.Is(err, tcase.expectedErr))
		})
	}
}

func TestOffice_Send_middleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	err := errors.New("interrupted")

	mw1, mw2, mw3 := mock_postdog.NewMockMiddleware(ctrl), mock_postdog.NewMockMiddleware(ctrl), mock_postdog.NewMockMiddleware(ctrl)
	mw1Call := mw1.EXPECT().Handle(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, let letter.Letter, next func(context.Context, letter.Letter) (letter.Letter, error)) (letter.Letter, error) {
			return next(ctx, let)
		})

	mw2.EXPECT().Handle(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, let letter.Letter, next func(context.Context, letter.Letter) (letter.Letter, error)) (letter.Letter, error) {
			return letter.Letter{}, err
		}).After(mw1Call)

	off := postdog.New(
		postdog.WithMiddleware(
			mw1,
			mw2,
			mw3,
		),
	)

	let := letter.Write(letter.Subject("Test"))
	trans := mock_postdog.NewMockTransport(ctrl)
	off.ConfigureTransport("test", trans)

	sendErr := off.Send(context.Background(), let)
	assert.True(t, errors.Is(sendErr, err))
}

func TestOffice_Send_errorlog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("send error")

	logger := mock_postdog.NewMockLogger(ctrl)
	logger.EXPECT().Log(expectedErr)

	trans := mock_postdog.NewMockTransport(ctrl)
	trans.EXPECT().Send(gomock.Any(), gomock.Any()).Return(expectedErr)

	off := postdog.New(postdog.WithLogger(logger))
	off.ConfigureTransport("test", trans)

	go off.Run(context.Background())

	err := off.Send(context.Background(), letter.Write())
	assert.True(t, errors.Is(err, expectedErr))

	<-time.After(time.Millisecond * 100)
}

func TestOffice_Send_beforeSendHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	let := letter.Write(letter.Subject("Test"))
	var wg sync.WaitGroup
	wg.Add(1)

	off := postdog.New(
		postdog.WithSendHook(postdog.BeforeSend, func(_ context.Context, hlet letter.Letter) {
			defer wg.Done()
			assert.Equal(t, let, hlet)
		}),
	)

	trans := mock_postdog.NewMockTransport(ctrl)
	trans.EXPECT().Send(context.Background(), let).Return(nil)
	off.ConfigureTransport("test", trans)

	err := off.Send(context.Background(), let)
	assert.Nil(t, err)

	wg.Wait()
}

func TestOffice_Send_afterSendHook(t *testing.T) {
	cases := map[string]struct {
		sendErr error
	}{
		"no error": {},
		"error": {
			sendErr: errors.New("send error"),
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			let := letter.Write(letter.Subject("Test"))
			var wg sync.WaitGroup
			wg.Add(1)

			off := postdog.New(
				postdog.WithSendHook(postdog.AfterSend, func(ctx context.Context, hlet letter.Letter) {
					defer wg.Done()
					assert.Equal(t, let, hlet)
					assert.Equal(t, tcase.sendErr, postdog.SendError(ctx))
				}),
			)

			trans := mock_postdog.NewMockTransport(ctrl)
			trans.EXPECT().Send(context.Background(), let).Return(tcase.sendErr)
			off.ConfigureTransport("test", trans)

			err := off.Send(context.Background(), let)
			assert.Equal(t, tcase.sendErr, err)

			wg.Wait()
		})
	}
}

func TestOffice_Dispatch(t *testing.T) {
	cases := map[string]struct {
		office      func(*gomock.Controller) *postdog.Office
		run         bool
		expectedErr error
	}{
		"unbuffered queue, not running": {
			office: func(ctrl *gomock.Controller) *postdog.Office {
				off := postdog.New()
				trans := mock_postdog.NewMockTransport(ctrl)
				off.ConfigureTransport("test", trans)
				return off
			},
			expectedErr: context.DeadlineExceeded,
		},
		"unbuffered queue, running": {
			office: func(ctrl *gomock.Controller) *postdog.Office {
				off := postdog.New()
				trans := mock_postdog.NewMockTransport(ctrl)
				trans.EXPECT().Send(
					gomock.Any(),
					gomock.AssignableToTypeOf(letter.Write()),
				)
				off.ConfigureTransport("test", trans)
				return off
			},
			run: true,
		},
		"buffered queue, not running": {
			office: func(ctrl *gomock.Controller) *postdog.Office {
				off := postdog.New(postdog.QueueBuffer(1))
				trans := mock_postdog.NewMockTransport(ctrl)
				off.ConfigureTransport("test", trans)
				return off
			},
		},
		"buffered queue, running": {
			office: func(ctrl *gomock.Controller) *postdog.Office {
				off := postdog.New(postdog.QueueBuffer(1))
				trans := mock_postdog.NewMockTransport(ctrl)
				trans.EXPECT().Send(
					gomock.Any(),
					gomock.AssignableToTypeOf(letter.Write()),
				)
				off.ConfigureTransport("test", trans)
				return off
			},
			run: true,
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			off := tcase.office(ctrl)

			if tcase.run {
				runCtx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go off.Run(runCtx)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()

			err := off.Dispatch(ctx, letter.Write())
			assert.True(t, errors.Is(err, tcase.expectedErr))

			// wait for queue
			<-time.After(time.Millisecond * 100)
		})
	}
}

func TestOffice_Dispatch_options(t *testing.T) {
	cases := map[string]struct {
		configure func(*postdog.Office, *gomock.Controller)
		options   []postdog.DispatchOption
	}{
		"unconfigured transport": {
			options: []postdog.DispatchOption{postdog.DispatchWith("test")},
		},
		"specify transport": {
			configure: func(off *postdog.Office, ctrl *gomock.Controller) {
				trans := mock_postdog.NewMockTransport(ctrl)
				trans.EXPECT().Send(gomock.Any(), gomock.AssignableToTypeOf(letter.Write()))
				off.ConfigureTransport("test1", mock_postdog.NewMockTransport(ctrl))
				off.ConfigureTransport("test2", trans)
				off.ConfigureTransport("test3", mock_postdog.NewMockTransport(ctrl))
			},
			options: []postdog.DispatchOption{postdog.DispatchWith("test2")},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			off := postdog.New(postdog.QueueBuffer(1))
			if tcase.configure != nil {
				tcase.configure(off, ctrl)
			}

			go off.Run(context.Background())

			assert.Nil(t, off.Dispatch(context.Background(), letter.Write(), tcase.options...))

			<-time.After(time.Millisecond * 100)
		})
	}
}

func TestOffice_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	off := postdog.New()
	trans := mock_postdog.NewMockTransport(ctrl)
	off.ConfigureTransport("test1", trans)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		assert.Nil(t, off.Run(ctx))
	}()

	<-time.After(time.Millisecond * 100)

	let1, let2, let3 := letter.Write(), letter.Write(), letter.Write()

	trans.EXPECT().Send(ctx, let1)
	trans.EXPECT().Send(ctx, let2)
	trans.EXPECT().Send(ctx, let3)

	assert.Nil(t, off.Dispatch(context.Background(), let1))
	assert.Nil(t, off.Dispatch(context.Background(), let2))
	assert.Nil(t, off.Dispatch(context.Background(), let3))

	<-time.After(time.Millisecond * 100)
}
