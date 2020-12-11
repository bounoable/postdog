package archive

import (
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMail(t *testing.T) {
	var m Mail
	assert.IsType(t, letter.Letter{}, m.Letter)
	assert.Equal(t, uuid.UUID{}, m.ID())
	assert.Equal(t, "", m.SendError())
	assert.Equal(t, time.Time{}, m.SentAt())
}

func TestExpandMail_basic(t *testing.T) {
	bm := basicMail{
		from: mail.Address{Name: "Bob Belcher", Address: "bob@example.com"},
		recipients: []mail.Address{
			{Name: "Linda Belcher", Address: "linda@example.com"},
			{Name: "Tina Belcher", Address: "tina@example.com"},
		},
		rfc: "rfc body",
	}

	m := ExpandMail(bm)
	assert.Equal(t, bm.from, m.From())
	assert.Equal(t, bm.recipients, m.Recipients())
	assert.Equal(t, bm.rfc, m.RFC())
}

func TestExpandMail_withID(t *testing.T) {
	idMail := Mail{id: uuid.New()}
	m := ExpandMail(idMail)
	assert.Equal(t, idMail.id, m.ID())
}

func TestExpandMail_withSendError(t *testing.T) {
	errMail := Mail{sendError: "send error"}
	m := ExpandMail(errMail)
	assert.Equal(t, errMail.sendError, m.SendError())
}

func TestExpandMail_withSendTime(t *testing.T) {
	timeMail := Mail{sentAt: time.Now()}
	m := ExpandMail(timeMail)
	assert.Equal(t, timeMail.sentAt, m.SentAt())
}

func TestMail_WithSendError(t *testing.T) {
	m := ExpandMail(letter.Write())
	err := errors.New("send error")
	m = m.WithSendError(err.Error())
	assert.Equal(t, "send error", m.SendError())
}

func TestMail_WithSendTime(t *testing.T) {
	m := ExpandMail(letter.Write())
	sa := time.Now()
	m = m.WithSendTime(sa)
	assert.Equal(t, sa, m.SentAt())
}

func TestMail_Map(t *testing.T) {
	mockSendError := errors.New("send error")
	mockSendTime := time.Now()

	tests := []struct {
		name    string
		give    Mail
		mapOpts []letter.MapOption
		want    func(Mail) map[string]interface{}
	}{
		{
			name: "default",
			give: ExpandMail(letter.Write(
				letter.From("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
				letter.CC("Tina Belcher", "tina@example.com"),
				letter.BCC("Gene Belcher", "gene@example.com"),
				letter.ReplyTo("Louise Belcher", "louise@example.com"),
				letter.Subject("Hi."),
				letter.Text("Hello"),
				letter.HTML("<p>Hello.</p>"),
				letter.Attach("attach1", []byte{1, 2, 3}, letter.ContentType("text/plain")),
			)).WithSendError(mockSendError.Error()).WithSendTime(mockSendTime),
			want: func(m Mail) map[string]interface{} {
				return map[string]interface{}{
					"from":        m.From(),
					"recipients":  m.Recipients(),
					"to":          m.To(),
					"cc":          m.CC(),
					"bcc":         m.BCC(),
					"replyTo":     m.ReplyTo(),
					"subject":     m.Subject(),
					"text":        m.Text(),
					"html":        m.HTML(),
					"attachments": mapAttachments(m.Attachments()),
					"sendError":   mockSendError.Error(),
					"sendTime":    mockSendTime.Format(time.RFC3339),
				}
			},
		},
		{
			name: "without contents",
			give: ExpandMail(letter.Write(
				letter.From("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
				letter.CC("Tina Belcher", "tina@example.com"),
				letter.BCC("Gene Belcher", "gene@example.com"),
				letter.ReplyTo("Louise Belcher", "louise@example.com"),
				letter.Subject("Hi."),
				letter.Text("Hello"),
				letter.HTML("<p>Hello.</p>"),
				letter.Attach("attach1", []byte{1, 2, 3}, letter.ContentType("text/plain")),
			)).WithSendError(mockSendError.Error()).WithSendTime(mockSendTime),
			mapOpts: []letter.MapOption{
				letter.WithoutAttachmentContent(),
			},
			want: func(m Mail) map[string]interface{} {
				return map[string]interface{}{
					"from":        m.From(),
					"recipients":  m.Recipients(),
					"to":          m.To(),
					"cc":          m.CC(),
					"bcc":         m.BCC(),
					"replyTo":     m.ReplyTo(),
					"subject":     m.Subject(),
					"text":        m.Text(),
					"html":        m.HTML(),
					"attachments": mapAttachments(m.Attachments(), letter.WithoutAttachmentContent()),
					"sendError":   mockSendError.Error(),
					"sendTime":    mockSendTime.Format(time.RFC3339),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(tt.give), tt.give.Map(tt.mapOpts...))
		})
	}
}

type basicMail struct {
	from       mail.Address
	recipients []mail.Address
	rfc        string
}

func (m basicMail) From() mail.Address {
	return m.from
}

func (m basicMail) Recipients() []mail.Address {
	return m.recipients
}

func (m basicMail) RFC() string {
	return m.rfc
}

func mapAttachments(ats []letter.Attachment, opts ...letter.MapOption) []map[string]interface{} {
	mapped := make([]map[string]interface{}, len(ats))
	for i, at := range ats {
		mapped[i] = at.Map(opts...)
	}
	return mapped
}
