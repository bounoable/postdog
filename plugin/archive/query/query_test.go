package query_test

import (
	"net/mail"
	"testing"
	"time"

	"github.com/bounoable/postdog/plugin/archive/query"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []query.Option
		want query.Query
	}{
		{
			name: "From()",
			opts: []query.Option{
				query.From(mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}),
				query.From(mail.Address{Name: "Linda Belcher", Address: "linda@example.com"}),
			},
			want: query.Query{
				From: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "To()",
			opts: []query.Option{
				query.To(mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}),
				query.To(mail.Address{Name: "Linda Belcher", Address: "linda@example.com"}),
			},
			want: query.Query{
				To: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "CC()",
			opts: []query.Option{
				query.CC(mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}),
				query.CC(mail.Address{Name: "Linda Belcher", Address: "linda@example.com"}),
			},
			want: query.Query{
				CC: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "BCC()",
			opts: []query.Option{
				query.BCC(mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}),
				query.BCC(mail.Address{Name: "Linda Belcher", Address: "linda@example.com"}),
			},
			want: query.Query{
				BCC: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "Recipients()",
			opts: []query.Option{
				query.Recipient(mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}),
				query.Recipient(mail.Address{Name: "Linda Belcher", Address: "linda@example.com"}),
			},
			want: query.Query{
				Recipients: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "Subject()",
			opts: []query.Option{
				query.Subject("Subject 1", "Subject 2"),
				query.Subject("Subject 3", "Subject 4"),
			},
			want: query.Query{
				Subjects: []string{"Subject 1", "Subject 2", "Subject 3", "Subject 4"},
			},
		},
		{
			name: "Text()",
			opts: []query.Option{
				query.Text("text body 1", "text body 2"),
				query.Text("text body 3", "text body 4"),
			},
			want: query.Query{
				Texts: []string{"text body 1", "text body 2", "text body 3", "text body 4"},
			},
		},
		{
			name: "HTML()",
			opts: []query.Option{
				query.HTML("html body 1", "html body 2"),
				query.HTML("html body 3", "html body 4"),
			},
			want: query.Query{
				HTML: []string{"html body 1", "html body 2", "html body 3", "html body 4"},
			},
		},
		{
			name: "RFC()",
			opts: []query.Option{
				query.RFC("rfc body 1", "rfc body 2"),
				query.RFC("rfc body 3", "rfc body 4"),
			},
			want: query.Query{
				RFC: []string{"rfc body 1", "rfc body 2", "rfc body 3", "rfc body 4"},
			},
		},
		{
			name: "AttachmentFilename()",
			opts: []query.Option{
				query.AttachmentFilename("attach1", "attach2"),
				query.AttachmentFilename("attach3", "attach4"),
			},
			want: query.Query{
				Attachment: query.AttachmentFilter{
					Filenames: []string{"attach1", "attach2", "attach3", "attach4"},
				},
			},
		},
		{
			name: "AttachmentSize()",
			opts: []query.Option{
				query.AttachmentSize(123, 234, 345),
				query.AttachmentSize(456, 567, 678),
			},
			want: query.Query{
				Attachment: query.AttachmentFilter{
					Size: query.AttachmentSizeFilter{
						Exact: []int{123, 234, 345, 456, 567, 678},
					},
				},
			},
		},
		{
			name: "AttachmentSizeRange()",
			opts: []query.Option{
				query.AttachmentSizeRange(0, 1500),
				query.AttachmentSizeRange(40, 28172),
			},
			want: query.Query{
				Attachment: query.AttachmentFilter{
					Size: query.AttachmentSizeFilter{
						Ranges: [][2]int{{0, 1500}, {40, 28172}},
					},
				},
			},
		},
		{
			name: "AttachmentContentType()",
			opts: []query.Option{
				query.AttachmentContentType("text/plain", "text/html"),
				query.AttachmentContentType("application/json", "application/octet-stream"),
			},
			want: query.Query{
				Attachment: query.AttachmentFilter{
					ContentTypes: []string{
						"text/plain", "text/html",
						"application/json", "application/octet-stream",
					},
				},
			},
		},
		{
			name: "AttachmentContent()",
			opts: []query.Option{
				query.AttachmentContent([]byte{1, 2, 3}),
				query.AttachmentContent([]byte{4, 5, 6}, []byte{7, 8, 9}),
			},
			want: query.Query{
				Attachment: query.AttachmentFilter{
					Contents: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
				},
			},
		},
		{
			name: "SendError()",
			opts: []query.Option{
				query.SendError("send error 1", "send error 2"),
				query.SendError("send error 3", "send error 4"),
			},
			want: query.Query{
				SendErrors: []string{"send error 1", "send error 2", "send error 3", "send error 4"},
			},
		},
		{
			name: "SendTime()",
			opts: []query.Option{
				query.SendTime(
					time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				),
				query.SendTime(
					time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 4, 0, 0, 0, 0, time.UTC),
				),
			},
			want: query.Query{
				SendTimes: []time.Time{
					time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 4, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "Sort(): default",
			want: query.Query{
				Sorting:       query.SortAny,
				SortDirection: query.SortAsc,
			},
		},
		{
			name: "Sort(): SendDate (asc)",
			opts: []query.Option{
				query.Sort(query.SortSendTime, query.SortAsc),
			},
			want: query.Query{
				Sorting:       query.SortSendTime,
				SortDirection: query.SortAsc,
			},
		},
		{
			name: "Sort(): SendDate (desc)",
			opts: []query.Option{
				query.Sort(query.SortSendTime, query.SortDesc),
			},
			want: query.Query{
				Sorting:       query.SortSendTime,
				SortDirection: query.SortDesc,
			},
		},
		{
			name: "Paginate()",
			opts: []query.Option{
				query.Paginate(3, 40),
			},
			want: query.Query{
				Pagination: query.Pagination{
					Page:    3,
					PerPage: 40,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, query.New(tt.opts...))
		})
	}
}
