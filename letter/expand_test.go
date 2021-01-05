package letter

import (
	"bytes"
	"net/mail"
	"net/textproto"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/stretchr/testify/assert"
)

type basicMail struct {
	from       mail.Address
	recipients []mail.Address
	body       string
}

type attachmentMail struct {
	basicMail
	attachments []attachment
}

type attachment struct {
	name string
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
			give: aBasicMail,
			want: Letter{
				from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
				recipients: []mail.Address{
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "letter with custom RFC body",
			give: Write(
				From("Bob Belcher", "bob@example.com"),
				RFC("rfc body"),
			),
			want: Letter{
				from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
				rfc:  "rfc body",
			},
		},
		{
			name: "mail with Attachments() method",
			give: attachmentMail{
				attachments: []attachment{{"foo"}, {"bar"}, {"foobar"}},
			},
			want: Write(
				AttachReader("foo", bytes.NewReader([]byte{1, 2, 3}), AttachmentType("application/pdf")),
				AttachReader("bar", bytes.NewReader([]byte{1, 2, 3}), AttachmentType("application/pdf")),
				AttachReader("foobar", bytes.NewReader([]byte{1, 2, 3}), AttachmentType("application/pdf")),
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

func (m basicMail) RFC() string {
	return m.body
}

func (m attachmentMail) Attachments() []attachment {
	return m.attachments
}

func (a attachment) Filename() string {
	return a.name
}

func (a attachment) Content() []byte {
	return []byte{1, 2, 3}
}

func (a attachment) ContentType() string {
	return "application/pdf"
}

func (a attachment) Header() textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"foo": []string{"bar"},
	}
}

var _ postdog.Mail = basicMail{}
var _ postdog.Mail = attachmentMail{}
var _ interface {
	Filename() string
	Content() []byte
	ContentType() string
	Header() textproto.MIMEHeader
} = attachment{}
