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
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
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

			Convey("Given an archive Plugin that uses that Store", func() {
				a := archive.New(s)

				Convey("Given a Postdog that uses the archive", func() {
					tr := newMockTransport(ctrl)
					dog := postdog.New(
						postdog.WithTransport("test", tr),
						a,
					)

					Convey("When I send a Mail and the Transport doesn't fail", WithStoreInsert(t, s, WithTransportSend(tr, func() {
						err := dog.Send(context.Background(), mockLetter)
						<-time.After(5 * time.Millisecond) // wait because hooks get called concurrently

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})
					})))

					Convey("When I send a Mail and the Transport fails", WithStoreErrorInsert(t, s, WithTransportError(tr, func() {
						err := dog.Send(context.Background(), mockLetter)
						<-time.After(5 * time.Millisecond)

						Convey("It should fail", func() {
							So(errors.Is(err, mockTransportError), ShouldBeTrue)
						})
					})))
				})
			})

			Convey("Given an archive Plugin with a Logger that uses that Store", func() {
				l := mock_archive.NewMockPrinter(ctrl)
				a := archive.New(s, archive.WithLogger(l))

				Convey("Given a Postdog that uses the archive", func() {
					tr := mock_postdog.NewMockTransport(ctrl)
					dog := postdog.New(
						postdog.WithTransport("test", tr),
						a,
					)

					Convey("When I send a Mail then the Store insert fails", WithFailingStoreInsert(s, WithTransportSend(tr, func() {
						l.EXPECT().Print(gomock.Any()).DoAndReturn(func(v ...interface{}) {
							assert.Equal(t, fmt.Errorf("store: %w", mockInsertError), v[0])
						})

						err := dog.Send(context.Background(), mockLetter)
						<-time.After(5 * time.Millisecond)

						Convey("It should't fail fail", func() {
							So(err, ShouldBeNil)
						})
					})))

					Convey("When I send a Mail and the Transport fails, then the Store insert fails", WithFailingStoreErrorInsert(s, WithTransportError(tr, func() {
						l.EXPECT().Print(gomock.Any()).DoAndReturn(func(v ...interface{}) {
							assert.Equal(t, fmt.Errorf("store: %w", mockInsertError), v[0])
						})

						err := dog.Send(context.Background(), mockLetter)
						<-time.After(5 * time.Millisecond)

						Convey("It should fail", func() {
							So(errors.Is(err, mockTransportError), ShouldBeTrue)
						})
					})))
				})
			})
		})
	})
}

func newMockTransport(ctrl *gomock.Controller) *mock_postdog.MockTransport {
	tr := mock_postdog.NewMockTransport(ctrl)
	return tr
}

func WithStoreInsert(t *testing.T, s *mock_archive.MockStore, fn func()) func() {
	return func() {
		s.EXPECT().Insert(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *archive.Mail) error {
			assert.Equal(t, archive.NewMail(mockLetter), m)
			return nil
		})
		fn()
	}
}

func WithStoreErrorInsert(t *testing.T, s *mock_archive.MockStore, fn func()) func() {
	return func() {
		s.EXPECT().Insert(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *archive.Mail) error {
			assert.Equal(t, archive.NewMailWithError(mockLetter, mockTransportError), m)
			return nil
		})
		fn()
	}
}

func WithFailingStoreInsert(s *mock_archive.MockStore, fn func()) func() {
	return func() {
		m := archive.NewMail(mockLetter)
		s.EXPECT().Insert(gomock.Any(), m).Return(mockInsertError)
		fn()
	}
}

func WithFailingStoreErrorInsert(s *mock_archive.MockStore, fn func()) func() {
	return func() {
		m := archive.NewMailWithError(mockLetter, mockTransportError)
		s.EXPECT().Insert(gomock.Any(), m).Return(mockInsertError)
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
