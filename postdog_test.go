package postdog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	mock_postdog "github.com/bounoable/postdog/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockLetter, _ = letter.Write(
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
		letter.Subject("Hi."),
		letter.Content("Hello.", "<p>Hello.</p>"),
	)
)

func TestPostdog(t *testing.T) {
	Convey("Postdog", t, func() {
		ctrl := gomock.NewController(t)
		Reset(ctrl.Finish)

		Convey("Feature: Send mail", func() {
			Convey("Scenario: No transport configured", func() {
				dog := postdog.New()

				Convey("When I send a mail without specifying the transport", func() {
					err := dog.Send(context.Background(), mockLetter)

					Convey("An error should be returned", func() {
						So(err, ShouldBeError, postdog.ErrNoTransport)
					})
				})

				Convey("When I send a mail and specify the transport", func() {
					err := dog.Send(context.Background(), mockLetter, postdog.Use("test"))

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

					err := dog.Send(context.Background(), mockLetter)

					Convey("Then the configured transport should be used", func() {
						So(err, ShouldBeNil)
					})
				})

				Convey("When I send a mail and specify the transport", func() {
					tr.EXPECT().
						Send(gomock.Any(), mockLetter).
						Return(nil)

					err := dog.Send(context.Background(), mockLetter, postdog.Use("test"))

					Convey("Then the configured transport should be used", func() {
						So(err, ShouldBeNil)
					})
				})

				Convey("When I send a mail and specify another transport", func() {
					err := dog.Send(context.Background(), mockLetter, postdog.Use("test2"))

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

					err := dog.Send(context.Background(), mockLetter)

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

						err := dog.Send(context.Background(), mockLetter)

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
					return letter.Must(letter.Write(
						letter.ToAddress(m.Recipients()...),
						letter.To("Linda Belcher", "linda@example.com"),
					))
				})

				mw2 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					return letter.Must(letter.Write(
						letter.ToAddress(m.Recipients()...),
						letter.To("Tina Belcher", "tina@example.com"),
					))
				})

				mw3 := newMockMiddleware(ctrl, func(m postdog.Mail) postdog.Mail {
					return letter.Must(letter.Write(
						letter.ToAddress(m.Recipients()...),
						letter.To("Gene Belcher", "gene@example.com"),
					))
				})

				tr := mock_postdog.NewMockTransport(ctrl)
				tr.EXPECT().
					Send(gomock.Any(), letter.Must(letter.Write(
						letter.To("Linda Belcher", "linda@example.com"),
						letter.To("Tina Belcher", "tina@example.com"),
						letter.To("Gene Belcher", "gene@example.com"),
					))).
					Return(nil)

				dog := postdog.New(
					postdog.WithTransport("test", tr),
					postdog.WithMiddleware(mw1, mw2, mw3),
				)

				Convey("When I send a mail", func() {
					err := dog.Send(context.Background(), letter.Must(letter.Write()))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
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
			ctx context.Context,
			m postdog.Mail,
			next func(context.Context, postdog.Mail) (postdog.Mail, error),
		) (postdog.Mail, error) {
			return next(ctx, fn(m))
		})
	return mw
}
