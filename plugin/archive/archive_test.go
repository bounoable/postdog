package archive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	mock_postdog "github.com/bounoable/postdog/mocks"
	"github.com/bounoable/postdog/plugin/archive"
	mock_archive "github.com/bounoable/postdog/plugin/archive/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockLetter = letter.Write(
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
	)
	mockTransportError = errors.New("transport error")
	mockInsertError    = errors.New("insert error")
)

func TestNew(t *testing.T) {
	Convey("Archive", t, func() {
		ctrl := gomock.NewController(t)
		Reset(ctrl.Finish)

		Convey("Given a Store", func() {
			s := mock_archive.NewMockStore(ctrl)

			Convey("Given an archive Plugin that uses the Store", func() {
				a := archive.New(s)
				tr := newMockTransport(ctrl)

				Convey("Given a Transport that doesn't fail to send", WithTransportSend(tr, func() {
					Convey("Given a Postdog that uses the archive and Transport", func() {
						dog := postdog.New(
							postdog.WithTransport("test", tr),
							a,
						)

						Convey("When I send a Mail", WithStoreInsert(s, func(storedMail <-chan postdog.Mail) {
							err := dog.Send(context.Background(), mockLetter)

							Convey("It shouldn't fail", func() {
								<-storedMail
								So(err, ShouldBeNil)
							})

							Convey("The stored mail should be expandable to the sent mail", func() {
								m := archive.ExpandMail(<-storedMail)
								So(m.Letter, ShouldResemble, mockLetter)
							})

							Convey("The stored mail should have a non-zero uuid", func() {
								m := archive.ExpandMail(<-storedMail)
								So(m.ID(), ShouldHaveSameTypeAs, uuid.Nil)
								So(m.ID(), ShouldNotEqual, uuid.Nil)
							})
						}))
					})
				}))

				Convey("Given a Transport that fails to send after 50 milliseconds", WithDelayedTransportError(tr, 50*time.Millisecond, func() {
					Convey("Given a Postdog that uses the archive and Transport", func() {
						dog := postdog.New(
							postdog.WithTransport("test", tr),
							a,
						)

						Convey("When I send a Mail", WithStoreInsert(s, func(storedMail <-chan postdog.Mail) {
							start := time.Now()
							err := dog.Send(context.Background(), mockLetter)

							Convey("It should fail", func() {
								So(errors.Is(err, mockTransportError), ShouldBeTrue)
							})

							Convey("The stored mail should be expandable to the sent mail", func() {
								m := archive.ExpandMail(<-storedMail)
								So(m.Letter, ShouldResemble, mockLetter)
							})

							Convey("The stored mail should contain the send error", func() {
								pm := <-storedMail
								m := archive.ExpandMail(pm)
								So(m.SendError(), ShouldContainSubstring, mockTransportError.Error())
							})

							Convey("The stored mail should contain the send time", func() {
								pm := <-storedMail
								m := archive.ExpandMail(pm)
								So(m.SentAt().UnixNano(), ShouldAlmostEqual, start.Add(50*time.Millisecond).UnixNano(), 10*time.Millisecond)
							})
						}))
					})
				}))
			})

			Convey("Given that the Store fails to insert mails", WithFailingStoreInsert(s, func() {
				Convey("Given an archive Plugin with a logger that uses the Store", func() {
					logger := make(loggerChan)
					a := archive.New(s, archive.WithLogger(logger))

					tr := newMockTransport(ctrl)
					Convey("Given a Transport that doesn't fail to send", WithTransportSend(tr, func() {
						Convey("Given a Postdog that uses the archive and Transport", func() {
							dog := postdog.New(
								postdog.WithTransport("test", tr),
								a,
							)

							Convey("When I Send a Mail", func() {
								err := dog.Send(context.Background(), mockLetter)

								Convey("It shouldn't fail", func() {
									So(err, ShouldBeNil)
								})

								Convey("The insert error should be logged", func() {
									So(<-logger, ShouldEqual, fmt.Sprintf("Failed to insert mail into store: %s\n", mockInsertError.Error()))
								})
							})
						})
					}))
				})
			}))
		})
	})
}

func newMockTransport(ctrl *gomock.Controller) *mock_postdog.MockTransport {
	tr := mock_postdog.NewMockTransport(ctrl)
	return tr
}

func WithStoreInsert(s *mock_archive.MockStore, fn func(<-chan postdog.Mail)) func() {
	return func() {
		ch := make(chan postdog.Mail, 1)
		s.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, pm postdog.Mail) error {
				ch <- pm
				return nil
			})
		fn(ch)
	}
}

func WithFailingStoreInsert(s *mock_archive.MockStore, fn func()) func() {
	return func() {
		s.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(mockInsertError)
		fn()
	}
}

func WithTransportSend(tr *mock_postdog.MockTransport, fn func()) func() {
	return func() {
		tr.EXPECT().Send(gomock.Any(), mockLetter).Return(nil)
		fn()
	}
}

func WithTransportError(tr *mock_postdog.MockTransport, fn func()) func() {
	return func() {
		tr.EXPECT().Send(gomock.Any(), mockLetter).Return(mockTransportError)
		fn()
	}
}

func WithDelayedTransportError(tr *mock_postdog.MockTransport, delay time.Duration, fn func()) func() {
	return func() {
		tr.EXPECT().
			Send(gomock.Any(), mockLetter).
			DoAndReturn(func(context.Context, postdog.Mail) error {
				<-time.After(delay)
				return mockTransportError
			})
		fn()
	}
}

type loggerChan chan string

func (lc loggerChan) Print(v ...interface{}) {
	lc <- fmt.Sprint(v...)
}
