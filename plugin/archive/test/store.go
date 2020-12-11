package test

import (
	"context"
	stdctx "context"
	"errors"
	"fmt"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/archive"
	"github.com/bounoable/postdog/plugin/archive/query"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockMail = archive.ExpandMail(letter.Write(
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
		letter.CC("Tina Belcher", "tina@example.com"),
		letter.BCC("Gene Belcher", "gene@example.com"),
		letter.Subject("Hi."),
		letter.Content("Hello.", "<p>Hello.</p>"),
		letter.Attach("attach-1", []byte{1}),
	))

	errMockSend = errors.New("mock send error")
)

// Store tests the archive.Store returned by newStore.
func Store(t *testing.T, newStore func() archive.Store) {
	Convey("Store", t, func() {
		Convey("Insert()", func() {
			s := newStore()

			Convey("When I insert a mail", func() {
				err := s.Insert(stdctx.Background(), mockMail)

				Convey("It shouldn't fail", func() {
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("Find()", func() {
			Convey("Given a Store", func() {
				s := newStore()

				Convey("When I try to find a mail that doesn't exist", func() {
					m, err := s.Find(stdctx.Background(), uuid.New())

					Convey("It should fail with archive.ErrNotFound", func() {
						So(errors.Is(err, archive.ErrNotFound), ShouldBeTrue)
					})

					Convey("The mail should be zero-value", func() {
						So(m, ShouldResemble, archive.Mail{})
					})
				})

				Convey("When I insert a mail with an ID", func() {
					id := uuid.New()
					m := mockMail.WithID(id)
					err := s.Insert(stdctx.Background(), m)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("When I call Find() with the mail's ID", func() {
						found, err := s.Find(stdctx.Background(), id)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("It should find the correct mail", func() {
							So(found, ShouldResemble, m)
						})
					})
				})
			})
		})

		Convey("Query()", func() {
			Convey("Given a Store with 3 mails", withFilledStore(newStore, 3, func(s archive.Store, mockMails []archive.Mail) {
				Convey("When I query the sender `Sender 3 <sender3@example.com>`", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.From(mail.Address{
							Name:    "Sender 3",
							Address: "sender3@example.com",
						}),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[2])
					})
				})

				Convey("When I query the recipient `Recipient 2 <rcpt2@example.com>`", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.Recipient(mail.Address{
							Name:    "Recipient 2",
							Address: "rcpt2@example.com",
						}),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[1])
					})
				})

				Convey("When I query the `To` recipient `Recipient 2 <rcpt2@example.com>`", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.To(mail.Address{
							Name:    "Recipient 2",
							Address: "rcpt2@example.com",
						}),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[1])
					})
				})

				Convey("When I query the `Cc` recipient `CC Recipient 3 <ccrcpt3@example.com>`", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.CC(mail.Address{
							Name:    "CC Recipient 3",
							Address: "ccrcpt3@example.com",
						}),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[2])
					})
				})

				Convey("When I query the `Bcc` recipient `BCC Recipient 1 <ccrcpt1@example.com>`", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.BCC(mail.Address{
							Name:    "BCC Recipient 1",
							Address: "bccrcpt1@example.com",
						}),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[0])
					})
				})

				Convey("When I query the RFC body of a mail", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.RFC(`To: "Recipient 2" <rcpt2@example.com>`),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[1])
					})
				})

				Convey("When I query the text body of a mail", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.Text("Content 2"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[1])
					})
				})

				Convey("When I query the HTML body of a mail", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.HTML("<p>Content 2"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[1])
					})
				})

				Convey("When I query the subject of a mail", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.Subject("Subject 2"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[1])
					})
				})

				Convey("When I query for attachments by filename", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.AttachmentFilename("Attachment 3"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have one element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[2])
					})
				})

				Convey("When I query for attachments by file size", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.AttachmentSize(3),
						query.AttachmentSize(10),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have 2 elements", func() {
						So(drain(cur), ShouldHaveLength, 2)
					})

					Convey("Cursor should return the mails", func() {
						mails := drain(cur)
						So(mails, ShouldResemble, mockMails[:len(mockMails)-1])
					})
				})

				Convey("When I query for attachments by file size range", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.AttachmentSizeRange(3, 10),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have 2 elements", func() {
						So(drain(cur), ShouldHaveLength, 2)
					})

					Convey("Cursor should return the mails", func() {
						mails := drain(cur)
						So(mails, ShouldResemble, mockMails[:len(mockMails)-1])
					})
				})

				Convey("When I query for attachments by content type", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.AttachmentContentType("text/html"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have 1 element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mails := drain(cur)
						mail := mails[0]
						So(mail, ShouldResemble, mockMails[2])
					})
				})

				Convey("When I query for attachments by file content", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.AttachmentContent(
							[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
							[]byte{1, 2, 3},
						),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have 2 elements", func() {
						So(drain(cur), ShouldHaveLength, 2)
					})

					Convey("Cursor should return the correct mails", func() {
						mails := drain(cur)
						So(mails[0], ShouldResemble, mockMails[0])
						So(mails[1], ShouldResemble, mockMails[1])
					})
				})

				testSorting(s, mockMails)
			}))

			Convey("Given a Store with 5 failed mails", withFilledErrMailStore(newStore, 5, func(s archive.Store) {
				Convey("When I query for the full error message", func() {
					cur, err := s.Query(context.Background(), query.New(
						query.SendError(errMockSend.Error()+" 2"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have 1 element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mail := drain(cur)[0]
						So(mail.SendError(), ShouldEqual, errMockSend.Error()+" 2")
					})
				})

				Convey("When I query for an error message substring", func() {
					cur, err := s.Query(context.Background(), query.New(
						query.SendError("2"),
					))

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("Cursor should have 1 element", func() {
						So(drain(cur), ShouldHaveLength, 1)
					})

					Convey("Cursor should return the correct mail", func() {
						mail := drain(cur)[0]
						So(mail.SendError(), ShouldEqual, errMockSend.Error()+" 2")
					})
				})
			}))
		})

		Convey("Remove()", func() {
			Convey("Given a Store", func() {
				s := newStore()

				Convey("When I insert a Mail", func() {
					m := mockMail.WithID(uuid.New())
					err := s.Insert(stdctx.Background(), m)

					Convey("It shouldn't fail", func() {
						So(err, ShouldBeNil)
					})

					Convey("When I remove the Mail", func() {
						err := s.Remove(stdctx.Background(), m)

						Convey("It shouldn't fail", func() {
							So(err, ShouldBeNil)
						})

						Convey("When I try to find that Mail", func() {
							m, err := s.Find(stdctx.Background(), m.ID())

							Convey("It should fail with archive.ErrNotFound", func() {
								So(errors.Is(err, archive.ErrNotFound), ShouldBeTrue)
							})

							Convey("The returned Mail should be zero-value", func() {
								So(m, ShouldResemble, archive.Mail{})
							})
						})
					})
				})
			})
		})
	})
}

func testSorting(s archive.Store, mockMails []archive.Mail) {
	sortings := []query.Sorting{
		query.SortSendTime,
		query.SortSubject,
	}

	compareFns := map[query.Sorting]map[query.SortDirection]func([]archive.Mail, int, int) bool{
		query.SortSendTime: {
			query.SortAsc: func(mails []archive.Mail, a, b int) bool {
				return mails[a].SentAt().Before(mails[b].SentAt())
			},
			query.SortDesc: func(mails []archive.Mail, a, b int) bool {
				return mails[a].SentAt().After(mails[b].SentAt())
			},
		},
		query.SortSubject: {
			query.SortAsc: func(mails []archive.Mail, a, b int) bool {
				return mails[a].Subject() < mails[b].Subject()
			},
			query.SortDesc: func(mails []archive.Mail, a, b int) bool {
				return mails[a].Subject() > mails[b].Subject()
			},
		},
	}

	for _, sorting := range sortings {
		Convey(fmt.Sprintf("When I sort a Query by %s (ascending)", sorting), func() {
			cur, err := s.Query(stdctx.Background(), query.New(
				query.Sort(sorting, query.SortAsc),
			))

			Convey("It shouldn't fail", func() {
				So(err, ShouldBeNil)
			})

			Convey("Cursor should have 3 elements", func() {
				So(drain(cur), ShouldHaveLength, 3)
			})

			Convey("Cursor should return the mails in correct order", func() {
				want := mockMails
				sort.Slice(want, func(a, b int) bool {
					return compareFns[sorting][query.SortAsc](want, a, b)
				})

				mails := drain(cur)
				So(mails, ShouldResemble, want)
			})
		})

		Convey(fmt.Sprintf("When I sort a Query by %s (descending)", sorting), func() {
			cur, err := s.Query(stdctx.Background(), query.New(
				query.Sort(sorting, query.SortDesc),
			))

			Convey("It shouldn't fail", func() {
				So(err, ShouldBeNil)
			})

			Convey("Cursor should have 3 elements", func() {
				So(drain(cur), ShouldHaveLength, 3)
			})

			Convey("Cursor should return the mails in correct order", func() {
				want := mockMails
				sort.Slice(want, func(a, b int) bool {
					return compareFns[sorting][query.SortDesc](want, a, b)
				})

				mails := drain(cur)
				So(mails, ShouldResemble, want)
			})
		})
	}
}

func makeMails(count int) []archive.Mail {
	mails := make([]archive.Mail, count)
	for i := 0; i < count; i++ {
		var contentType string
		var size int
		switch i % 3 {
		case 0:
			contentType = "application/octet-stream"
			size = 3
		case 1:
			contentType = "text/plain"
			size = 10
		case 2:
			contentType = "text/html"
			size = 20
		}
		content := make([]byte, size)
		for i := range content {
			content[i] = byte(i + 1)
		}
		mails[i] = archive.ExpandMail(letter.Write(
			letter.From(fmt.Sprintf("Sender %d", i+1), fmt.Sprintf("sender%d@example.com", i+1)),
			letter.To(fmt.Sprintf("Recipient %d", i+1), fmt.Sprintf("rcpt%d@example.com", i+1)),
			letter.CC(fmt.Sprintf("CC Recipient %d", i+1), fmt.Sprintf("ccrcpt%d@example.com", i+1)),
			letter.BCC(fmt.Sprintf("BCC Recipient %d", i+1), fmt.Sprintf("bccrcpt%d@example.com", i+1)),
			letter.Subject(fmt.Sprintf("Subject %d", i+1)),
			letter.Content(fmt.Sprintf("Content %d", i+1), fmt.Sprintf("<p>Content %d</p>", i+1)),
			letter.Attach(fmt.Sprintf("Attachment %d", i+1), content, letter.ContentType(contentType)),
		)).WithSendTime(time.Now().Add(time.Duration(i) * time.Minute))
	}
	return mails
}

func withFilledStore(newStore func() archive.Store, count int, fn func(archive.Store, []archive.Mail)) func() {
	return func() {
		s := newStore()
		mails := makeMails(count)
		for _, m := range mails {
			if err := s.Insert(stdctx.Background(), m); err != nil {
				panic(err)
			}
		}
		fn(s, mails)
	}
}

func withFilledErrMailStore(newStore func() archive.Store, count int, fn func(archive.Store)) func() {
	return func() {
		s := newStore()
		mails := makeMails(count)
		for i := range mails {
			mails[i] = mails[i].WithSendError(fmt.Sprintf("%s %d", errMockSend.Error(), i+1))
			if err := s.Insert(stdctx.Background(), mails[i]); err != nil {
				panic(err)
			}
		}
		fn(s)
	}
}

func drain(cur archive.Cursor) []archive.Mail {
	mails, err := cur.All(stdctx.Background())
	if err != nil {
		panic(err)
	}
	return mails
}
