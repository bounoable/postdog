package archive

import (
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/stretchr/testify/assert"
)

func TestMail(t *testing.T) {
	var m Mail
	assert.IsType(t, letter.Letter{}, m.Letter)
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
