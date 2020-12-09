package archive

import (
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
)

// Mail is the archived form of a sent mail, containing the send time and send error of the mail.
type Mail struct {
	letter.Letter

	sentAt    time.Time
	sendError string
}

// SentAt returns the time at which the mail was sent.
func (m Mail) SentAt() time.Time {
	return m.sentAt
}

// SendError returns the message of the send error. An empty string means there was no error.
func (m Mail) SendError() string {
	return m.sendError
}

// ExpandMail takes a postdog.Mail and builds a Mail from it. If pm has a
// SendError() method, the error will be added to the Mail. If pm has a
// SentAt() method, the time will be added as the send time.
func ExpandMail(pm postdog.Mail) Mail {
	m := Mail{
		Letter: letter.Expand(pm),
	}

	if errMail, ok := pm.(interface{ SendError() string }); ok {
		m.sendError = errMail.SendError()
	}

	if timeMail, ok := pm.(interface{ SentAt() time.Time }); ok {
		m.sentAt = timeMail.SentAt()
	}

	return m
}
