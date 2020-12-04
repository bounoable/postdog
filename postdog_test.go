package postdog_test

import (
	"context"
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
