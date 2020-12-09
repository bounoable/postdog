package test

import (
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

	mockMails = makeMails(3)
)

// Store runs the basic store functionality against the archive.Store returned by newStore.
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

		Convey("Query()", func() {
			s := newStore()

			Convey("Given a Store with 3 mails", withFilledStore(s, func() {
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

						So(mail.From(), ShouldResemble, mockMails[2].From())
						So(mail.Recipients(), ShouldResemble, mockMails[2].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[2].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[2].From())
						So(mail.Recipients(), ShouldResemble, mockMails[2].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[2].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[0].From())
						So(mail.Recipients(), ShouldResemble, mockMails[0].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[0].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
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

						So(mail.From(), ShouldResemble, mockMails[2].From())
						So(mail.Recipients(), ShouldResemble, mockMails[2].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[2].RFC())
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
						for i := 0; i < len(mockMails)-1; i++ {
							mail := mails[i]
							So(mail.From(), ShouldResemble, mockMails[i].From())
							So(mail.Recipients(), ShouldResemble, mockMails[i].Recipients())
							So(mail.RFC(), ShouldEqual, mockMails[i].RFC())
						}
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
						for i := 0; i < len(mockMails)-1; i++ {
							mail := mails[i]
							So(mail.From(), ShouldResemble, mockMails[i].From())
							So(mail.Recipients(), ShouldResemble, mockMails[i].Recipients())
							So(mail.RFC(), ShouldEqual, mockMails[i].RFC())
						}
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

						So(mail.From(), ShouldResemble, mockMails[2].From())
						So(mail.Recipients(), ShouldResemble, mockMails[2].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[2].RFC())
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

						mail := mails[0]
						So(mail.From(), ShouldResemble, mockMails[0].From())
						So(mail.Recipients(), ShouldResemble, mockMails[0].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[0].RFC())

						mail = mails[1]
						So(mail.From(), ShouldResemble, mockMails[1].From())
						So(mail.Recipients(), ShouldResemble, mockMails[1].Recipients())
						So(mail.RFC(), ShouldEqual, mockMails[1].RFC())
					})
				})

				Convey("When I sort a Query by descending send time", func() {
					cur, err := s.Query(stdctx.Background(), query.New(
						query.Sort(query.SortSendTime, query.SortDesc),
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
							return want[a].SentAt().
								After(want[b].SentAt())
						})

						mails := drain(cur)
						for i := 0; i < len(want); i++ {
							mail := mails[i]
							So(mail.From(), ShouldResemble, want[i].From())
							So(mail.Recipients(), ShouldResemble, want[i].Recipients())
							So(mail.RFC(), ShouldEqual, want[i].RFC())
						}
					})
				})
			}))
		})
	})
}

// ExtendedStore runs the full store functionality against the archive.Store returned by newStore.
func ExtendedStore(t *testing.T, newStore func() archive.Store) {
	Store(t, newStore)
	extendedStore(t, newStore)
}

func extendedStore(t *testing.T, newStore func() archive.Store) {
	// Convey("ExtendedStore", t, func() {
	// 	Convey("Query()", func() {
	// 		Convey("Given a context with a send error", func() {
	// 			sendError := errors.New("send error")
	// 			ctx := context.WithSendError(stdctx.Background(), sendError)

	// 			Convey("When I insert a mail with that context", func() {

	// 			})
	// 		})
	// 	})
	// })
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

func withFilledStore(s archive.Store, fn func()) func() {
	return func() {
		for _, m := range mockMails {
			if err := s.Insert(stdctx.Background(), m); err != nil {
				panic(err)
			}
		}
		fn()
	}
}

func drain(cur archive.Cursor) []archive.Mail {
	mails, err := cur.All(stdctx.Background())
	if err != nil {
		panic(err)
	}
	return mails
}
