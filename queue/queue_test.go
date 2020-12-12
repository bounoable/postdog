package queue_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/internal/testing/should"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/queue"
	"github.com/bounoable/postdog/queue/dispatch"
	mock_queue "github.com/bounoable/postdog/queue/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var mockLetter = letter.Write(
	letter.From("Bob Belcher", "bob@example.com"),
	letter.To("Linda Belcher", "linda@example.com"),
	letter.Content("Hello.", "<p>Hello.</p>"),
)

var mockError = errors.New("mock error")

func TestQueue(t *testing.T) {
	Convey("Dispatch()", t, func() {
		ctrl := gomock.NewController(t)
		Reset(ctrl.Finish)

		Convey("Given a Mailer that takes 50 milliseconds to send a mail", WithDelayedMailer(ctrl, 50*time.Millisecond, func(m *mock_queue.MockMailer) {
			Convey("Given an unbuffered, started *Queue that uses the Mailer", func() {
				q := queue.New(m)
				q.Start()

				Convey("When I dispatch a mail", func() {
					job, err := q.Dispatch(context.Background(), mockLetter)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("The job should know the dispatch time", func() {
						So(job.DispatchedAt().UnixNano(), ShouldAlmostEqual, time.Now().UnixNano(), time.Millisecond)
					})

					Convey("job.Err() should return nil", func() {
						So(job.Err(), ShouldBeNil)
					})

					Convey("job.Runtime() should return the elapsed time", func() {
						<-time.After(time.Microsecond * 500)
						So(job.Runtime(), ShouldAlmostEqual, time.Now().Sub(job.DispatchedAt()), time.Microsecond*500)
					})

					Convey("job.Done() should return a channel that has no values and is not closed", func() {
						d := job.Done()

						So(d, ShouldNotBeNil)
						So(d, should.BeOpen)
					})

					Convey("successive calls to job.Done() should return the same channel", func() {
						d1 := job.Done()
						d2 := job.Done()
						d3 := job.Done()

						So(d1, ShouldEqual, d2)
						So(d2, ShouldEqual, d3)
					})

					Convey("job.Done() should be closed after ~50 milliseconds", func() {
						<-time.After(60 * time.Millisecond)
						So(job.Done(), should.BeClosed)
					})

					Convey("When I cancel the job before it's done", func() {
						err := job.Cancel(context.Background())

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("job.Done() should be closed", func() {
							So(job.Done(), should.BeClosed)
						})

						Convey("job.Err() should return queue.ErrCanceled", func() {
							So(errors.Is(job.Err(), queue.ErrCanceled), ShouldBeTrue)
						})
					})

					Convey("When I cancel the job after it's done", func() {
						<-time.After(time.Millisecond * 70)
						err := job.Cancel(context.Background())

						Convey("It should fail", func() {
							So(errors.Is(err, queue.ErrFinished), ShouldBeTrue)
						})

						Convey("job.Done() should be closed", func() {
							So(job.Done(), should.BeClosed)
						})

						Convey("job.Err() should return nil", func() {
							So(job.Err(), ShouldBeNil)
						})
					})

					Convey("When I cancel the job using a canceled context", func() {
						ctx, cancel := context.WithCancel(context.Background())
						cancel()
						err := job.Cancel(ctx)

						Convey("It should fail with context.Canceled", func() {
							So(errors.Is(err, context.Canceled), ShouldBeTrue)
						})

						Convey("job.Done() should not be closed", func() {
							So(job.Done(), should.BeOpen)
						})

						Convey("job.Err() should return nil", func() {
							So(job.Err(), ShouldBeNil)
						})
					})

					Convey("When I dispatch another mail", func() {
						start := time.Now()
						job2, err2 := q.Dispatch(context.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err2, ShouldBeNil)
						})

						Convey("It should take ~50 milliseconds to dispatch", func() {
							end := time.Now()
							dur := end.Sub(start)

							So(dur, ShouldAlmostEqual, time.Millisecond*50, 10*time.Millisecond)
							So(job2.DispatchedAt().UnixNano(), ShouldAlmostEqual, end.UnixNano(), 10*time.Millisecond)
						})
					})

					Convey("When I dispatch another mail with a context that's canceled before the first job is finished", func() {
						ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*25)
						Reset(cancel)

						job2, err2 := q.Dispatch(ctx, mockLetter)

						Convey("It should fail with the context error", func() {
							So(err2, ShouldNotBeNil)
							So(errors.Is(err2, ctx.Err()), ShouldBeTrue)
							So(job2, ShouldBeNil)
						})
					})

					Convey("When I dispatch another mail with a timeout of 20 milliseconds", func() {
						job2, err2 := q.Dispatch(context.Background(), mockLetter, dispatch.Timeout(time.Millisecond*20))

						Convey("It should fail with context.DeadlineExceeded", func() {
							So(errors.Is(err2, context.DeadlineExceeded), ShouldBeTrue)
						})

						Convey("Job should be nil", func() {
							So(job2, ShouldBeNil)
						})
					})
				})
			})

			Convey("Given a buffered (1), started *Queue that uses that Mailer", func() {
				q := queue.New(m, queue.Buffer(1))
				q.Start()

				Convey("When I dispatch a mail", func() {
					start := time.Now()
					_, err := q.Dispatch(context.Background(), mockLetter)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("When I dispatch a second mail", func() {
						_, err2 := q.Dispatch(context.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err2, ShouldBeNil)
						})

						Convey("When I dispatch a third mail", func() {
							job3, err3 := q.Dispatch(context.Background(), mockLetter)

							Convey("It shouldn't fail", func() {
								So(err3, ShouldBeNil)
							})

							Convey("It should take ~50 milliseconds for the third mail to dispatch", func() {
								end := time.Now()
								dur := end.Sub(start)

								So(dur, ShouldAlmostEqual, 50*time.Millisecond, time.Millisecond*5)
								So(job3.DispatchedAt().UnixNano(), ShouldAlmostEqual, end.UnixNano(), time.Millisecond*5)
							})
						})
					})
				})
			})

			Convey("Given an unstarted *Queue that used that Mailer", func() {
				q := queue.New(m)

				Convey("When I dispatch a mail", func() {
					job, err := q.Dispatch(context.Background(), mockLetter)

					Convey("It should fail with queue.ErrNotStarted", func() {
						So(err, ShouldNotBeNil)
						So(errors.Is(err, queue.ErrNotStarted), ShouldBeTrue)
					})

					Convey("Job should be nil", func() {
						So(job, ShouldBeNil)
					})
				})

				Convey("When I start the queue", func() {
					err := q.Start()

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("It should be started", func() {
						So(q.Started(), ShouldBeTrue)
					})

					Convey("When I try to start the queue again", func() {
						err := q.Start()

						Convey("It should fail with queue.ErrStarted", func() {
							So(err, ShouldBeError)
							So(errors.Is(err, queue.ErrStarted), ShouldBeTrue)
						})
					})

					Convey("When I stop the queue", func() {
						err := q.Stop(context.Background())

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("It should be stopped", func() {
							So(q.Started(), ShouldBeFalse)
						})

						Convey("When I try stop the queue again", func() {
							err := q.Stop(context.Background())

							Convey("It should fail with queue.ErrNotStarted", func() {
								So(errors.Is(err, queue.ErrNotStarted), ShouldBeTrue)
							})
						})
					})

					Convey("When I stop the queue with a canceled context", func() {
						ctx, cancel := context.WithCancel(context.Background())
						cancel()
						err := q.Stop(ctx)

						Convey("It should fail with the context error", func() {
							So(err, ShouldNotBeNil)
							So(errors.Is(err, ctx.Err()), ShouldBeTrue)
						})
					})
				})
			})

			Convey("Given a started *Queue with 2 workers that uses that Mailer", func() {
				q := queue.New(m, queue.Workers(2))
				q.Start()

				Convey("When I dispatch a mail", func() {
					job, err := q.Dispatch(context.Background(), mockLetter)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("When I dispatch another mail", func() {
						job2, err2 := q.Dispatch(context.Background(), mockLetter)

						Convey("It shouldn't fail", func() {
							So(err2, ShouldBeNil)
						})

						Convey("It should be dispatched at roughly the same time as the first mail", func() {
							So(job2.DispatchedAt().UnixNano(), ShouldAlmostEqual, job.DispatchedAt().UnixNano(), time.Millisecond)
						})
					})
				})
			})
		}))

		Convey("Given a Mailer that fails to send mails", WithErrorMailer(ctrl, func(m *mock_queue.MockMailer) {
			Convey("Given a started *Queue that uses that Mailer", func() {
				q := queue.New(m)
				q.Start()

				Convey("When I dispatch a mail", func() {
					job, err := q.Dispatch(context.Background(), mockLetter)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("job.Done() should be closed", func() {
						<-time.After(time.Microsecond * 100)
						So(job.Done(), should.BeClosed)
					})

					Convey("job.Err() should return Mailer error", func() {
						<-time.After(time.Microsecond * 100)
						So(errors.Is(job.Err(), mockError), ShouldBeTrue)
					})
				})
			})
		}))

		Convey("Given a Mailer that counts postdog.SendOptions passed to it", WithOptionCountingMailer(ctrl, func(m *mock_queue.MockMailer, optCount <-chan int) {
			Convey("Given a started *Queue that uses that Mailer", func() {
				q := queue.New(m)
				q.Start()

				Convey("When I dispatch a mail with 2 postdog.SendOptions", func() {
					opts := []postdog.SendOption{
						postdog.Use("a"),
						postdog.Use("b"),
					}
					job, err := q.Dispatch(context.Background(), mockLetter, dispatch.SendOptions(opts...))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("<-optCount should be 2", func() {
						So(<-optCount, ShouldEqual, 2)
					})

					Convey("job.SendOptions() should return the passed options", func() {
						So(job.SendOptions(), ShouldHaveLength, 2)
					})
				})
			})
		}))
	})
}

func WithDelayedMailer(ctrl *gomock.Controller, d time.Duration, fn func(*mock_queue.MockMailer)) func() {
	return func() {
		m := mock_queue.NewMockMailer(ctrl)
		m.EXPECT().
			Send(gomock.Any(), mockLetter).
			DoAndReturn(func(ctx context.Context, _ postdog.Mail) error {
				select {
				case <-ctx.Done():
					return fmt.Errorf("abort send: %w", ctx.Err())
				case <-time.After(d):
					return nil
				}
			}).
			AnyTimes()
		fn(m)
	}
}

func WithErrorMailer(ctrl *gomock.Controller, fn func(*mock_queue.MockMailer)) func() {
	return func() {
		m := mock_queue.NewMockMailer(ctrl)
		m.EXPECT().
			Send(gomock.Any(), mockLetter).
			DoAndReturn(func(context.Context, postdog.Mail) error {
				return mockError
			}).
			AnyTimes()
		fn(m)
	}
}

func WithOptionCountingMailer(ctrl *gomock.Controller, fn func(*mock_queue.MockMailer, <-chan int)) func() {
	return func() {
		count := make(chan int)
		m := mock_queue.NewMockMailer(ctrl)
		m.EXPECT().
			Send(gomock.Any(), mockLetter, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ postdog.Mail, opts ...postdog.SendOption) error {
				count <- len(opts)
				return nil
			})
		fn(m, count)
	}
}
