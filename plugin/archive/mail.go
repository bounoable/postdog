package archive

import (
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/letter/expand"
	"github.com/bounoable/postdog/letter/mapper"
	"github.com/google/uuid"
)

// Mail is the archived form of a sent mail, containing the send time and send error of the mail.
type Mail struct {
	letter.Letter

	id        uuid.UUID
	sentAt    time.Time
	sendError string
}

// ExpandMail takes a postdog.Mail and builds a Mail from it. If pm has a
// SendError() method, the error will be added to the Mail. If pm has a
// SentAt() method, the time will be added as the send time.
func ExpandMail(pm postdog.Mail, opts ...expand.Option) Mail {
	if m, ok := pm.(Mail); ok {
		return m
	}

	m := Mail{Letter: letter.Expand(pm, opts...)}

	if idMail, ok := pm.(interface{ ID() uuid.UUID }); ok {
		m.id = idMail.ID()
	}

	if errMail, ok := pm.(interface{ SendError() string }); ok {
		m.sendError = errMail.SendError()
	}

	if timeMail, ok := pm.(interface{ SentAt() time.Time }); ok {
		m.sentAt = timeMail.SentAt()
	}

	return m
}

// ID returns the mail's ID.
func (m Mail) ID() uuid.UUID {
	return m.id
}

// SentAt returns the time at which the mail was sent.
func (m Mail) SentAt() time.Time {
	return m.sentAt
}

// SendError returns the message of the send error. An empty string means there was no error.
func (m Mail) SendError() string {
	return m.sendError
}

// WithSendError returns a copy of m with it's send error set to err.
func (m Mail) WithSendError(err string) Mail {
	m.sendError = err
	return m
}

// WithSendTime returns a copy of m with it's send time set to t.
func (m Mail) WithSendTime(t time.Time) Mail {
	m.sentAt = t
	return m
}

// WithID returns a copy of m with it's ID set to id.
func (m Mail) WithID(id uuid.UUID) Mail {
	m.id = id
	return m
}

// Map maps m to a map[string]interface{}.
func (m Mail) Map(opts ...mapper.Option) map[string]interface{} {
	res := m.Letter.Map(opts...)
	res["id"] = m.id.String()
	res["sendError"] = m.sendError
	res["sentAt"] = m.sentAt.Format(time.RFC3339)
	return res
}

// Parse parses mm into m.
func (m *Mail) Parse(mm map[string]interface{}) {
	m.Letter.Parse(mm)
	if id, ok := mm["id"].(string); ok {
		if uid, err := uuid.Parse(id); err == nil {
			m.id = uid
		}
	}
	if sendError, ok := mm["sendError"].(string); ok {
		m.sendError = sendError
	}
	if sentAt, ok := mm["sentAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, sentAt); err == nil {
			m.sentAt = t.Round(0)
		}
	}
}
