package letter

import (
	"bytes"
	"net/mail"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter/rfc"
	"github.com/stretchr/testify/assert"
)

type basicMail struct {
	from       mail.Address
	recipients []mail.Address
	body       string
}

var (
	aBasicMail = basicMail{
		from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
		recipients: []mail.Address{
			{Name: "Linda Belcher", Address: "linda@example.com"},
		},
		body: "mail body",
	}
)

func TestExpand(t *testing.T) {
	tests := []struct {
		name string
		give postdog.Mail
		want Letter
	}{
		{
			name: "basic mail",
			give: aBasicMail,
			want: Letter{
				from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
				recipients: []mail.Address{
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
				rfc: "mail body",
			},
		},
		{
			name: "mail with To() method",
			give: Write(
				To("Bob Belcher", "bob@example.com"),
				To("Linda Belcher", "linda@example.com"),
			),
			want: Letter{
				to: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with CC() method",
			give: Write(
				CC("Bob Belcher", "bob@example.com"),
				CC("Linda Belcher", "linda@example.com"),
			),
			want: Letter{
				cc: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with BCC() method",
			give: Write(
				BCC("Bob Belcher", "bob@example.com"),
				BCC("Linda Belcher", "linda@example.com"),
			),
			want: Letter{
				bcc: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with ReplyTo() method",
			give: Write(
				ReplyTo("Bob Belcher", "bob@example.com"),
				ReplyTo("Linda Belcher", "linda@example.com"),
			),
			want: Letter{
				replyTo: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with Recipients() method",
			give: Write(
				Recipient("Bob Belcher", "bob@example.com"),
				Recipient("Linda Belcher", "linda@example.com"),
			),
			want: Letter{
				recipients: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with Recipients() and To() method with duplicates",
			give: Write(
				Recipient("Bob Belcher", "bob@example.com"),
				Recipient("Linda Belcher", "linda@example.com"),
				To("Linda Belcher", "linda@example.com"),
			),
			want: Letter{
				recipients: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				},
				to: []mail.Address{
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with Subject() method",
			give: Write(
				Subject("Hello."),
			),
			want: Letter{
				subject: "Hello.",
			},
		},
		{
			name: "mail with Text() method",
			give: Write(
				Text("Hello."),
			),
			want: Letter{
				text: "Hello.",
			},
		},
		{
			name: "mail with HTML() method",
			give: Write(
				HTML("<p>Hello.</p>"),
			),
			want: Letter{
				html: "<p>Hello.</p>",
			},
		},
		{
			name: "mail with RFC() method",
			give: Write(
				RFC("rfc body"),
			),
			want: Letter{
				rfc: "rfc body",
			},
		},
		{
			name: "mail with Attachments() method",
			give: Write(
				AttachReader("attach-1", bytes.NewReader([]byte{1, 2, 3})),
				AttachReader("attach-2", bytes.NewReader([]byte{2, 3, 4, 5})),
			),
			want: Write(
				AttachReader("attach-1", bytes.NewReader([]byte{1, 2, 3})),
				AttachReader("attach-2", bytes.NewReader([]byte{2, 3, 4, 5})),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Expand(tt.give))
		})
	}
}

func (m basicMail) From() mail.Address {
	return m.from
}

func (m basicMail) Recipients() []mail.Address {
	return m.recipients
}

func (m basicMail) RFC(...rfc.Option) string {
	return m.body
}

var _ postdog.Mail = basicMail{}
