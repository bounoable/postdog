package gmail_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/transport/gmail"
	mock_gmail "github.com/bounoable/postdog/transport/gmail/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	ggmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var mockLetter, _ = letter.Write(
	letter.From("Bob Belcher", "bob@example.com"),
	letter.To("Linda Belcher", "linda@example.com"),
	letter.Subject("Hi."),
	letter.Content("Hello.", "<p>Hello.</p>"),
)

func TestTransport(t *testing.T) {
	Convey("Transport", t, func() {
		ctrl := gomock.NewController(t)
		Reset(ctrl.Finish)

		Convey("Feature: Send", func() {
			Convey("Scenario: no credentials", func() {
				tr := gmail.Transport()

				Convey("When I send a mail", func() {
					err := tr.Send(context.Background(), mockLetter)

					Convey("An error should be returned", func() {
						So(errors.Is(err, gmail.ErrNoCredentials), ShouldBeTrue)
					})
				})
			})

			Convey("Scenario: provide sender directly", func() {
				sender := mock_gmail.NewMockSender(ctrl)

				// It should call the sender with the correct message
				expectSend(t, sender)

				tr := gmail.Transport(gmail.WithSender(sender))

				Convey("When I send a mail", func() {
					err := tr.Send(context.Background(), mockLetter)

					Convey("No error should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("Scenario: provide sender through token source", func() {
				tok := &oauth2.Token{}
				ts := oauth2.StaticTokenSource(tok)

				sender := mock_gmail.NewMockSender(ctrl)

				// It should call the sender with the correct message
				expectSend(t, sender)

				tr := gmail.Transport(
					gmail.WithTokenSource(ts),
					gmail.WithSenderFactory(func(_ context.Context, ts oauth2.TokenSource, _ ...option.ClientOption) (gmail.Sender, error) {
						return sender, nil
					}),
				)

				Convey("When I send a mail", func() {
					err := tr.Send(context.Background(), mockLetter)

					Convey("No error should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("Scenario: provide sender through token source factory", func() {
				tok := &oauth2.Token{}
				ts := oauth2.StaticTokenSource(tok)

				sender := mock_gmail.NewMockSender(ctrl)

				// It should call the sender with the correct message
				expectSend(t, sender)

				tr := gmail.Transport(
					gmail.WithTokenSourceFactory(func(context.Context, ...string) (oauth2.TokenSource, error) {
						return ts, nil
					}),
					gmail.WithSenderFactory(func(_ context.Context, ts oauth2.TokenSource, _ ...option.ClientOption) (gmail.Sender, error) {
						return sender, nil
					}),
				)

				Convey("When I send a mail", func() {
					err := tr.Send(context.Background(), mockLetter)

					Convey("No error should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})
}

func expectSend(t *testing.T, sender *mock_gmail.MockSender) {
	sender.EXPECT().
		Send("me", gomock.Any()).
		DoAndReturn(func(_ string, msg *ggmail.Message) error {
			expected := base64.URLEncoding.EncodeToString([]byte(mockLetter.RFC()))
			assert.Equal(t, expected, msg.Raw)
			return nil
		})
}
