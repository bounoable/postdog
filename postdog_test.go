package postdog_test

import (
	stdctx "context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	mock_postdog "github.com/bounoable/postdog/mocks"
	"github.com/bounoable/postdog/send"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
)

var (
	mockLetter = letter.Write(
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
		letter.Subject("Hi."),
		letter.Content("Hello.", "<p>Hello.</p>"),
	)

	mockError = errors.New("mock error")
)

func TestPostdog(t *testing.T) {
	Convey("Postdog", t, func() {
		ctrl := gomock.NewController(t)
		Reset(ctrl.Finish)

		Convey("Feature: Send mail", func() {
			Convey("Scenario: No transport configured", func() {
				dog := postdog.New()

				Convey("When I send a mail without specifying the transport", func() {
					err := dog.Send(stdctx.Background(), mockLetter)

					Convey("An error should be returned", func() {
						So(err, ShouldBeError, postdog.ErrNoTransport)
					})
				})

				Convey("When I send a mail and specify the transport", func() {
					err := dog.Send(stdctx.Background(), mockLetter, send.Use("test"))

					Convey("An error should be returned", func() {
						So(err, ShouldBeError, postdog.ErrUnconfiguredTransport)
					})
				})
			})

			Convey("Scenario: Transport configured", WithMockTransport(ctrl, func(tr *mock_postdog.MockTransport) {
				dog := postdog.New(
					postdog.WithTransport("test", tr),
				)

				Convey("When I send a mail without specifying the transport", func() {
					tr.EXPECT().
						Send(gomock.Any(), mockLetter).
						Return(nil)

					err := dog.Send(stdctx.Background(), mockLetter)

					Convey("Then the configured transport should be used", func() {
						So(err, ShouldBeNil)
					})
				})

				Convey("When I send a mail and specify the transport", func() {
					tr.EXPECT().
						Send(gomock.Any(), mockLetter).
						Return(nil)

					err := dog.Send(stdctx.Background(), mockLetter, send.Use("test"))

					Convey("Then the configured transport should be used", func() {
						So(err, ShouldBeNil)
					})
				})

				Convey("When I send a mail and specify another transport", func() {
					err := dog.Send(stdctx.Background(), mockLetter, send.Use("test2"))

					Convey("An error should be returned", func() {
						So(err, ShouldBeError, postdog.ErrUnconfiguredTransport)
					})
				})
			}))

			Convey("Scenario: Multiple transports configured", func() {
				tr1 := newMockTransport(ctrl)
				tr2 := newMockTransport(ctrl)
				tr3 := newMockTransport(ctrl)

				dog := postdog.New(
					postdog.WithTransport("test1", tr1),
					postdog.WithTransport("test2", tr2),
					postdog.WithTransport("test3", tr3),
				)

				Convey("When I send a mail without specifying the transport", func() {
					tr1.EXPECT().
						Send(gomock.Any(), mockLetter).
						Return(nil)

					err := dog.Send(stdctx.Background(), mockLetter)

					Convey("The first configured transport should be used", func() {
						So(err, ShouldBeNil)
					})
				})

				Convey("When I manually set the default transport", func() {
					dog.Use("test3")

					Convey("When I send a mail without specifying the transport", func() {
						tr3.EXPECT().
							Send(gomock.Any(), mockLetter).
							Return(nil)

						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("The configured default transport should be used", func() {
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})

		Convey("Feature: Get transport by name", func() {
			Convey("Given a *postdog.Dog with configured transports", func() {
				tr1 := mock_postdog.NewMockTransport(ctrl)
				tr2 := mock_postdog.NewMockTransport(ctrl)
				dog := postdog.New(
					postdog.WithTransport("test1", tr1),
					postdog.WithTransport("test2", tr2),
				)

				Convey("When I call dog.Transport() with the name of a configured transport", func() {
					tr, err := dog.Transport("test2")

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
						So(tr, ShouldEqual, tr2)
					})
				})

				Convey("When I call dog.Transport() with the name of an unconfigured transport", func() {
					tr, err := dog.Transport("test3")

					Convey("It should fail", func() {
						So(tr, ShouldBeNil)
						So(errors.Is(err, postdog.ErrUnconfiguredTransport), ShouldBeTrue)
					})
				})
			})
		})

		Convey("Feature: Middleware", func() {
			Convey("Given 3 middlewares that all add a recipient to the mail", func() {
				mw1 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					return letter.Write(
						letter.ToAddress(m.Recipients()...),
						letter.To("Linda Belcher", "linda@example.com"),
					)
				})

				mw2 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					return letter.Write(
						letter.ToAddress(m.Recipients()...),
						letter.To("Tina Belcher", "tina@example.com"),
					)
				})

				mw3 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					return letter.Write(
						letter.ToAddress(m.Recipients()...),
						letter.To("Gene Belcher", "gene@example.com"),
					)
				})

				tr := mock_postdog.NewMockTransport(ctrl)
				tr.EXPECT().
					Send(gomock.Any(), letter.Write(
						letter.To("Linda Belcher", "linda@example.com"),
						letter.To("Tina Belcher", "tina@example.com"),
						letter.To("Gene Belcher", "gene@example.com"),
					)).
					Return(nil)

				dog := postdog.New(
					postdog.WithTransport("test", tr),
					postdog.WithMiddleware(mw1, mw2, mw3),
				)

				Convey("When I send a mail", func() {
					err := dog.Send(stdctx.Background(), letter.Write())

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})

		Convey("Feature: Rate limiting", func() {
			Convey("Given a Transport", WithMockTransport(ctrl, func(tr *mock_postdog.MockTransport) {
				tr.EXPECT().
					Send(gomock.Any(), mockLetter).
					DoAndReturn(func(_ stdctx.Context, _ postdog.Mail) error {
						return nil
					}).
					AnyTimes()

				Convey("Given a rate of 2 mails per 50 milliseconds", func() {
					lim := rate.NewLimiter(rate.Limit((1000/50)*2), 1)
					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.WithRateLimiter(lim),
					)

					Convey("When I send a mail", func() {
						start := time.Now()
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("When I immediately send another mail", func() {
							err2 := dog.Send(stdctx.Background(), mockLetter)

							Convey("It shouldn't fail", func() {
								So(err2, ShouldBeNil)
							})

							Convey("When I immediately send a third mail", func() {
								err3 := dog.Send(stdctx.Background(), mockLetter)

								Convey("It shouldn't fail", func() {
									So(err3, ShouldBeNil)
								})

								Convey("It should take ~50 milliseconds for the third mail to be sent", func() {
									end := time.Now()
									dur := end.Sub(start)

									So(dur, ShouldAlmostEqual, 50*time.Millisecond, 8*time.Millisecond)
								})
							})
						})
					})
				})
			}))
		})

		Convey("Feature: Timeout", func() {
			Convey("Given a Transport that takes 50 milliseconds to send a mail", WithDelayedTransport(ctrl, 50*time.Millisecond, func(tr *mock_postdog.MockTransport) {
				dog := postdog.New(postdog.WithTransport("test", tr))

				Convey("When I send a mail with a timeout of 20 milliseconds", func() {
					err := dog.Send(stdctx.Background(), mockLetter, send.Timeout(20*time.Millisecond))

					Convey("It should fail with stdctx.DeadlineExceeded", func() {
						So(errors.Is(err, stdctx.DeadlineExceeded), ShouldBeTrue)
					})
				})

				Convey("When I send a mail with a context with a timeout of 20 milliseconds", func() {
					ctx, cancel := stdctx.WithTimeout(stdctx.Background(), 20*time.Millisecond)
					Reset(cancel)

					err := dog.Send(ctx, mockLetter)

					Convey("It should fail with stdctx.DeadlineExceeded", func() {
						So(errors.Is(err, stdctx.DeadlineExceeded), ShouldBeTrue)
					})
				})

				Convey("When I send a mail with a timeout of 60 milliseconds", func() {
					err := dog.Send(stdctx.Background(), mockLetter, send.Timeout(60*time.Millisecond))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})
				})
			}))
		})

		Convey("Feature: Hooks > BeforeSend", func() {
			Convey("Given a Transport that takes 50 milliseconds to send a Mail", WithDelayedTransport(ctrl, 50*time.Millisecond, func(tr *mock_postdog.MockTransport) {
				Convey("Given a single Hook", func() {
					handledAt := make(chan time.Time, 1)
					lis := mock_postdog.NewMockListener(ctrl)
					lis.EXPECT().
						Handle(gomock.Any(), postdog.BeforeSend, mockLetter).
						Do(func(stdctx.Context, postdog.Hook, postdog.Mail) {
							handledAt <- time.Now()
						})

					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.WithHook(postdog.BeforeSend, lis),
					)

					Convey("When I send a Mail", func() {
						start := time.Now()
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("The Hook Listener should have been called ~immediately", func() {
							So((<-handledAt).UnixNano(), ShouldAlmostEqual, start.UnixNano(), 500*time.Microsecond)
						})
					})
				})

				Convey("Given multiple Hooks that take ~50 milliseconds to execute", func() {
					// if runtime.GOMAXPROCS(0) < 2 {
					// 	// t.Skip("Skipping test because machine only runs on 1 CPU.")
					// 	return
					// }

					calls := make(chan time.Time, 3)
					lis1 := newDelayedListener(ctrl, 50*time.Millisecond, postdog.BeforeSend, calls)
					lis2 := newDelayedListener(ctrl, 50*time.Millisecond, postdog.BeforeSend, calls)
					lis3 := newDelayedListener(ctrl, 50*time.Millisecond, postdog.BeforeSend, calls)

					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.WithHook(postdog.BeforeSend, lis1),
						postdog.WithHook(postdog.BeforeSend, lis2),
						postdog.WithHook(postdog.BeforeSend, lis3),
					)

					Convey("When I send a Mail", func() {
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("Listeners should have been called", func() {
							called := make(chan struct{})
							go func() {
								defer close(called)
								<-calls
								<-calls
								<-calls
							}()
							select {
							case <-called:
							case <-time.After(20 * time.Millisecond):
								t.Fatal("listeners should have been called")
							}
						})

						Convey("Listeners should have been called concurrently", func() {
							t1, t2, t3 := <-calls, <-calls, <-calls
							So(t1.UnixNano(), ShouldAlmostEqual, t2.UnixNano(), 5*time.Millisecond)
							So(t2.UnixNano(), ShouldAlmostEqual, t3.UnixNano(), 5*time.Millisecond)
						})
					})
				})
			}))
		})

		Convey("Feature: Hooks > AfterSend", func() {
			Convey("Given a Transport that takes 50 milliseconds to send a Mail", WithDelayedTransport(ctrl, 50*time.Millisecond, func(tr *mock_postdog.MockTransport) {
				Convey("Given multiple Hooks that take ~50 milliseconds to execute", func() {
					calls := make(chan time.Time, 3)
					lis1 := newDelayedListener(ctrl, 50*time.Millisecond, postdog.AfterSend, calls)
					lis2 := newDelayedListener(ctrl, 50*time.Millisecond, postdog.AfterSend, calls)
					lis3 := newDelayedListener(ctrl, 50*time.Millisecond, postdog.AfterSend, calls)

					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.WithHook(postdog.AfterSend, lis1),
						postdog.WithHook(postdog.AfterSend, lis2),
						postdog.WithHook(postdog.AfterSend, lis3),
					)

					Convey("When I send a Mail", func() {
						start := time.Now()
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("Listeners should have been called after ~50 milliseconds", func() {
							<-time.After(55 * time.Millisecond)
							So(calls, ShouldHaveLength, 3)
						})

						Convey("Listeners should have been called concurrently", func() {
							t1, t2, t3 := <-calls, <-calls, <-calls
							So(t1.UnixNano(), ShouldAlmostEqual, t2.UnixNano(), 5*time.Millisecond)
							So(t2.UnixNano(), ShouldAlmostEqual, t3.UnixNano(), 5*time.Millisecond)
						})

						Convey("Listeners should have been called after send", func() {
							t1, t2, t3 := <-calls, <-calls, <-calls

							dur1 := t1.Sub(start)
							dur2 := t2.Sub(start)
							dur3 := t3.Sub(start)

							So(dur1, ShouldAlmostEqual, 50*time.Millisecond, 5*time.Millisecond)
							So(dur2, ShouldAlmostEqual, 50*time.Millisecond, 5*time.Millisecond)
							So(dur3, ShouldAlmostEqual, 50*time.Millisecond, 5*time.Millisecond)
						})
					})
				})

				Convey("Given a Listener that needs the send time of a Mail", func() {
					gotTime := make(chan time.Time, 1)
					lis := mock_postdog.NewMockListener(ctrl)
					lis.EXPECT().
						Handle(gomock.Any(), postdog.AfterSend, mockLetter).
						Do(func(ctx stdctx.Context, _ postdog.Hook, _ postdog.Mail) {
							gotTime <- postdog.SendTime(ctx)
						}).
						AnyTimes()

					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.WithHook(postdog.AfterSend, lis),
					)

					Convey("When I send a Mail", func() {
						start := time.Now()
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("The Listener should have received the send time", func() {
							So((<-gotTime).UnixNano(), ShouldAlmostEqual, start.Add(50*time.Millisecond).UnixNano(), 10*time.Millisecond)
						})
					})
				})
			}))

			Convey("Given a Transport that fails to send Mails", WithErrorTransport(ctrl, func(tr *mock_postdog.MockTransport) {
				Convey("Given a Listener that needs the send error", func() {
					gotError := make(chan error, 1)
					lis := mock_postdog.NewMockListener(ctrl)
					lis.EXPECT().
						Handle(gomock.Any(), postdog.AfterSend, mockLetter).
						Do(func(ctx stdctx.Context, _ postdog.Hook, _ postdog.Mail) {
							gotError <- postdog.SendError(ctx)
						}).
						AnyTimes()

					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.WithHook(postdog.AfterSend, lis),
					)

					Convey("When I send a Mail", func() {
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It should fail", func() {
							So(errors.Is(err, mockError), ShouldBeTrue)
						})

						Convey("The Listener should have received the error", func() {
							gotErr := <-gotError
							So(errors.Is(gotErr, mockError), ShouldBeTrue)
						})
					})
				})
			}))
		})

		Convey("Feature: Plugins", func() {
			Convey("Given some Options that are Middleware options", func() {
				var wg sync.WaitGroup
				wg.Add(3)
				mw1 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					wg.Done()
					return m
				})
				mw2 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					wg.Done()
					return m
				})
				mw3 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					wg.Done()
					return m
				})

				opts := []postdog.Option{
					postdog.WithMiddleware(mw1),
					postdog.WithMiddleware(mw2),
					postdog.WithMiddleware(mw3),
				}

				Convey("When I add them as a Plugin", func() {
					tr := newMockTransport(ctrl)
					tr.EXPECT().Send(gomock.Any(), mockLetter).Return(nil)

					dog := postdog.New(
						postdog.WithTransport("test", tr),
						postdog.Plugin(opts),
					)

					Convey("When I send a mail", func() {
						err := dog.Send(stdctx.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("All middlewares should have been called", func() {
							called := make(chan struct{})
							go func() {
								defer close(called)
								wg.Wait()
							}()

							select {
							case <-time.After(time.Millisecond * 10):
								t.Fatal("middlewares should have been called")
							case <-called:
							}
						})
					})
				})
			})
		})
	})
}

func WithMockTransport(ctrl *gomock.Controller, fn func(*mock_postdog.MockTransport)) func() {
	return func() {
		fn(newMockTransport(ctrl))
	}
}

func newMockTransport(ctrl *gomock.Controller) *mock_postdog.MockTransport {
	return mock_postdog.NewMockTransport(ctrl)
}

func newMockMiddleware(ctrl *gomock.Controller, fn func(postdog.Mail) postdog.Mail) postdog.Middleware {
	mw := mock_postdog.NewMockMiddleware(ctrl)
	mw.EXPECT().
		Handle(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx stdctx.Context,
			m postdog.Mail,
			next func(stdctx.Context, postdog.Mail) (postdog.Mail, error),
		) (postdog.Mail, error) {
			return next(ctx, fn(m))
		})
	return mw
}

func WithDelayedTransport(ctrl *gomock.Controller, delay time.Duration, fn func(*mock_postdog.MockTransport)) func() {
	return func() {
		tr := mock_postdog.NewMockTransport(ctrl)
		tr.EXPECT().
			Send(gomock.Any(), mockLetter).
			DoAndReturn(func(ctx stdctx.Context, _ postdog.Mail) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					return nil
				}
			})
		fn(tr)
	}
}

func WithErrorTransport(ctrl *gomock.Controller, fn func(*mock_postdog.MockTransport)) func() {
	return func() {
		tr := mock_postdog.NewMockTransport(ctrl)
		tr.EXPECT().
			Send(gomock.Any(), mockLetter).
			DoAndReturn(func(stdctx.Context, postdog.Mail) error {
				return mockError
			}).
			AnyTimes()
		fn(tr)
	}
}

func newDelayedListener(ctrl *gomock.Controller, delay time.Duration, h postdog.Hook, calls chan<- time.Time) *mock_postdog.MockListener {
	lis := mock_postdog.NewMockListener(ctrl)
	lis.EXPECT().Handle(gomock.Any(), h, mockLetter).Do(func(stdctx.Context, postdog.Hook, postdog.Mail) {
		calls <- time.Now()
		<-time.After(delay)
	}).AnyTimes()
	return lis
}
