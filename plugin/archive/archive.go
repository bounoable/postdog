package archive

import (
	"context"
	"fmt"
	"net/mail"
	"net/textproto"
	"reflect"

	"github.com/bounoable/postdog"
)

//go:generate mockgen -source=archive.go -destination=./mocks/archive.go

// Store is the underlying store for the Mails.
type Store interface {
	Insert(context.Context, *Mail) error
}

// A Mail represents an archived mail.
type Mail struct {
	from        mail.Address
	recipients  []mail.Address
	to, cc, bcc []mail.Address
	replyTo     []mail.Address
	text, html  string
	rfc         string
	attachments []Attachment
	sendError   string
}

// Attachment is a mail attachment.
type Attachment struct {
	filename    string
	content     []byte
	contentType string
	header      textproto.MIMEHeader
}

// Printer is the logger interface.
type Printer interface {
	Print(...interface{})
}

// Option is an archive option.
type Option func(*config)

type config struct {
	logger Printer
}

// New creates the archive plugin.
func New(s Store, opts ...Option) postdog.Plugin {
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	return postdog.Plugin{
		postdog.WithHook(postdog.AfterSend, postdog.ListenerFunc(func(
			ctx context.Context,
			_ postdog.Hook,
			pm postdog.Mail,
		) {
			sendError := postdog.SendError(ctx)

			var m *Mail
			if sendError != nil {
				m = NewMailWithError(pm, sendError)
			} else {
				m = NewMail(pm)
			}

			if err := s.Insert(ctx, m); err != nil {
				cfg.log(fmt.Errorf("store: %w", err))
			}
		})),
	}
}

// WithLogger returns an Option that sets the error logger.
func WithLogger(l Printer) Option {
	return func(cfg *config) {
		cfg.logger = l
	}
}

// NewMail creates a *Mail from a postdog.Mail.
//
// Add additional information
//
// If pm implements any of the optional methods To(), CC(), BCC(), ReplyTo(),
// Text(), HTML() or Attachments(), those methods will be called to retrieve
// the information which will be added to the returned *Mail.
//
// If pm has an Attachments() method, the return type of that method must be
// a slice of a type that implements the following methods: Filename() string,
// Content() []byte, ContentType() string, Header() textproto.MIMEHeader.
func NewMail(pm postdog.Mail) *Mail {
	m := Mail{
		from:       pm.From(),
		recipients: pm.Recipients(),
		rfc:        pm.RFC(),
	}

	if toMail, ok := pm.(interface{ To() []mail.Address }); ok {
		m.to = toMail.To()
	}

	if ccMail, ok := pm.(interface{ CC() []mail.Address }); ok {
		m.cc = ccMail.CC()
	}

	if bccMail, ok := pm.(interface{ BCC() []mail.Address }); ok {
		m.bcc = bccMail.BCC()
	}

	if rtMail, ok := pm.(interface{ ReplyTo() []mail.Address }); ok {
		m.replyTo = rtMail.ReplyTo()
	}

	if textMail, ok := pm.(interface{ Text() string }); ok {
		m.text = textMail.Text()
	}

	if htmlMail, ok := pm.(interface{ HTML() string }); ok {
		m.html = htmlMail.HTML()
	}

	m.attachments = getAttachments(pm)

	return &m
}

// NewMailWithError returns NewMail(pm), but adds sendError to the returned
// *Mail m, so that m.SendError() returns sendError.Error().
func NewMailWithError(pm postdog.Mail, sendError error) *Mail {
	m := NewMail(pm)
	if sendError != nil {
		m.sendError = sendError.Error()
	}
	return m
}

// From returns the sender of the mail.
func (m *Mail) From() mail.Address {
	return m.from
}

// Recipients returns the recipients of the mail.
func (m *Mail) Recipients() []mail.Address {
	return m.recipients
}

// To returns the `To` recipients of the mail.
func (m *Mail) To() []mail.Address {
	return m.to
}

// CC returns the `Cc` recipients of the mail.
func (m *Mail) CC() []mail.Address {
	return m.cc
}

// BCC returns the `Bcc` recipients of the mail.
func (m *Mail) BCC() []mail.Address {
	return m.bcc
}

// ReplyTo returns the `Reply-To` recipients of the mail.
func (m *Mail) ReplyTo() []mail.Address {
	return m.replyTo
}

// Text returns the text content of the mail.
func (m *Mail) Text() string {
	return m.text
}

// HTML returns the HTML content of the mail.
func (m *Mail) HTML() string {
	return m.html
}

// RFC returns the RFC 5322 representation of the mail.
func (m *Mail) RFC() string {
	return m.rfc
}

// Attachments returns the file attachments of the mail.
func (m *Mail) Attachments() []Attachment {
	return m.attachments
}

// SendError returns the send error message of the mail or an empty string if there was no error.
func (m *Mail) SendError() string {
	return m.sendError
}

func (cfg *config) log(v ...interface{}) {
	if cfg.logger != nil {
		cfg.logger.Print(v...)
	}
}

// getAttachments returns m.Attachments() as an []Attachment slice if the
// return type of m.Attachments() is a slice of a type that implements the
// methods Filename() string, Content() []byte, ContentType() string and
// Header() textproto.MIMEHeader.
func getAttachments(m postdog.Mail) []Attachment {
	type attachment interface {
		Filename() string
		Content() []byte
		ContentType() string
		Header() textproto.MIMEHeader
	}

	typ := reflect.TypeOf(m)
	val := reflect.ValueOf(m)

	method, ok := typ.MethodByName("Attachments")
	if !ok {
		return nil
	}

	methodType := method.Type
	if methodType.NumOut() != 1 {
		return nil
	}

	outType := methodType.Out(0)
	if outType.Kind() != reflect.Slice {
		return nil
	}

	outSliceType := outType.Elem()

	attachType := reflect.TypeOf(new(attachment)).Elem()
	if !outSliceType.Implements(attachType) {
		return nil
	}

	ret := method.Func.Call([]reflect.Value{val})
	if len(ret) == 0 {
		return nil
	}

	len := ret[0].Len()
	if len == 0 {
		return nil
	}

	result := make([]Attachment, len)

	for i := 0; i < len; i++ {
		at := ret[0].Index(i).Interface().(attachment)
		result[i] = Attachment{
			filename:    at.Filename(),
			content:     at.Content(),
			contentType: at.ContentType(),
			header:      at.Header(),
		}
	}

	return result
}
