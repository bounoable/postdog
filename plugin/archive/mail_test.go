package archive

import (
	"bytes"
	"errors"
	"net/mail"
	"net/textproto"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/stretchr/testify/assert"
)

var (
	aBasicMail = basicMail{
		from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
		recipients: []mail.Address{
			{Name: "Linda Belcher", Address: "linda@example.com"},
		},
		body: "mail body",
	}
)

func TestNewMail(t *testing.T) {
	tests := []struct {
		name string
		give postdog.Mail
		want *Mail
	}{
		{
			name: "basic mail",
			give: aBasicMail,
			want: &Mail{
				from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
				recipients: []mail.Address{
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
				rfc: "mail body",
			},
		},
		{
			name: "mail with To() method",
			give: letter.Write(
				letter.To("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
			),
			want: &Mail{
				recipients: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
				rfc: letter.Write(
					letter.To("Bob Belcher", "bob@example.com"),
					letter.To("Linda Belcher", "linda@example.com"),
				).RFC(),
				to: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with CC() method",
			give: letter.Write(
				letter.CC("Bob Belcher", "bob@example.com"),
				letter.CC("Linda Belcher", "linda@example.com"),
			),
			want: &Mail{
				recipients: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
				rfc: letter.Write(
					letter.CC("Bob Belcher", "bob@example.com"),
					letter.CC("Linda Belcher", "linda@example.com"),
				).RFC(),
				cc: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with BCC() method",
			give: letter.Write(
				letter.BCC("Bob Belcher", "bob@example.com"),
				letter.BCC("Linda Belcher", "linda@example.com"),
			),
			want: &Mail{
				recipients: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
				rfc: letter.Write(
					letter.BCC("Bob Belcher", "bob@example.com"),
					letter.BCC("Linda Belcher", "linda@example.com"),
				).RFC(),
				bcc: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with ReplyTo() method",
			give: letter.Write(
				letter.ReplyTo("Bob Belcher", "bob@example.com"),
				letter.ReplyTo("Linda Belcher", "linda@example.com"),
			),
			want: &Mail{
				rfc: letter.Write(
					letter.ReplyTo("Bob Belcher", "bob@example.com"),
					letter.ReplyTo("Linda Belcher", "linda@example.com"),
				).RFC(),
				replyTo: []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				},
			},
		},
		{
			name: "mail with Text() method",
			give: letter.Write(
				letter.Text("Hello."),
			),
			want: &Mail{
				rfc: letter.Write(
					letter.Text("Hello."),
				).RFC(),
				text: "Hello.",
			},
		},
		{
			name: "mail with HTML() method",
			give: letter.Write(
				letter.HTML("<p>Hello.</p>"),
			),
			want: &Mail{
				rfc: letter.Write(
					letter.HTML("<p>Hello.</p>"),
				).RFC(),
				html: "<p>Hello.</p>",
			},
		},
		{
			name: "mail with Attachments() method",
			give: letter.Write(
				letter.AttachReader("attach-1", bytes.NewReader([]byte{1, 2, 3})),
				letter.AttachReader("attach-2", bytes.NewReader([]byte{2, 3, 4, 5})),
			),
			want: &Mail{
				rfc: letter.Write(
					letter.AttachReader("attach-1", bytes.NewReader([]byte{1, 2, 3})),
					letter.AttachReader("attach-2", bytes.NewReader([]byte{2, 3, 4, 5})),
				).RFC(),
				attachments: []Attachment{
					{filename: "attach-1", content: []byte{1, 2, 3}, contentType: "application/octet-stream", header: letter.Write(
						letter.AttachReader("attach-1", bytes.NewReader([]byte{1, 2, 3})),
					).Attachments()[0].Header()},
					{filename: "attach-2", content: []byte{2, 3, 4, 5}, contentType: "application/octet-stream", header: letter.Write(
						letter.AttachReader("attach-2", bytes.NewReader([]byte{2, 3, 4, 5})),
					).Attachments()[0].Header()},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMail(tt.give))
		})
	}
}

func TestMail_From(t *testing.T) {
	addr := mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}
	m := &Mail{from: addr}
	assert.Equal(t, addr, m.From())
}

func TestMail_Recipients(t *testing.T) {
	rcpts := []mail.Address{
		{Name: "Bob Belcher", Address: "bob@example.com"},
		{Name: "Linda Belcher", Address: "linda@example.com"},
	}
	m := &Mail{recipients: rcpts}
	assert.Equal(t, rcpts, m.Recipients())
}

func TestMail_To(t *testing.T) {
	to := []mail.Address{
		{Name: "Bob Belcher", Address: "bob@example.com"},
		{Name: "Linda Belcher", Address: "linda@example.com"},
	}
	m := &Mail{to: to}
	assert.Equal(t, to, m.To())
}

func TestMail_CC(t *testing.T) {
	cc := []mail.Address{
		{Name: "Bob Belcher", Address: "bob@example.com"},
		{Name: "Linda Belcher", Address: "linda@example.com"},
	}
	m := &Mail{cc: cc}
	assert.Equal(t, cc, m.CC())
}

func TestMail_BCC(t *testing.T) {
	bcc := []mail.Address{
		{Name: "Bob Belcher", Address: "bob@example.com"},
		{Name: "Linda Belcher", Address: "linda@example.com"},
	}
	m := &Mail{bcc: bcc}
	assert.Equal(t, bcc, m.BCC())
}

func TestMail_ReplyTo(t *testing.T) {
	replyTo := []mail.Address{
		{Name: "Bob Belcher", Address: "bob@example.com"},
		{Name: "Linda Belcher", Address: "linda@example.com"},
	}
	m := &Mail{replyTo: replyTo}
	assert.Equal(t, replyTo, m.ReplyTo())
}

func TestMail_Text(t *testing.T) {
	text := "text body"
	m := &Mail{text: text}
	assert.Equal(t, text, m.Text())
}

func TestMail_HTML(t *testing.T) {
	html := "html body"
	m := &Mail{html: html}
	assert.Equal(t, html, m.HTML())
}

func TestMail_Attachments(t *testing.T) {
	ats := []Attachment{
		{filename: "file-a", content: []byte{1, 2, 3}, contentType: "application/octet-stream", header: textproto.MIMEHeader{
			"header-a": []string{"value-a", "value-b"},
		}},
		{filename: "file-b", content: []byte{4, 5, 6, 7}, contentType: "text/plain", header: textproto.MIMEHeader{
			"header-a": []string{"value-a"},
		}},
	}
	m := &Mail{attachments: ats}
	assert.Equal(t, ats, m.Attachments())
}

func TestMail_RFC(t *testing.T) {
	rfc := "rfc body"
	m := &Mail{rfc: rfc}
	assert.Equal(t, rfc, m.RFC())
}

func TestMail_SendError(t *testing.T) {
	err := errors.New("send error")
	m := &Mail{sendError: err.Error()}
	assert.Equal(t, err.Error(), m.SendError())
}

type basicMail struct {
	from       mail.Address
	recipients []mail.Address
	body       string
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

var _ postdog.Mail = basicMail{}

type mailWithTo struct {
	basicMail
	to []mail.Address
}

func (m mailWithTo) To() []mail.Address {
	return m.to
}
