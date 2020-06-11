package smtp

import (
	"bytes"
	"context"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"gopkg.in/mail.v2"
)

type transport struct {
	dialer *mail.Dialer
}

// NewTransport initializes an SMTP transport.
func NewTransport(host string, port int, username, password string) office.Transport {
	return NewTransportDialer(mail.NewDialer(host, port, username, password))
}

// NewTransportDialer ...
func NewTransportDialer(dialer *mail.Dialer) office.Transport {
	return transport{
		dialer: dialer,
	}
}

func (trans transport) Send(ctx context.Context, let *letter.Letter) error {
	msg := mail.NewMessage(
		mail.SetEncoding(mail.Base64),
		mail.SetCharset("utf-8"),
	)

	msg.SetHeader("From", let.From.String())
	if len(let.To) > 0 {
		msg.SetHeader("To", letter.RecipientsHeader(let.To))
	}

	if len(let.CC) > 0 {
		msg.SetHeader("Cc", letter.RecipientsHeader(let.CC))
	}

	if len(let.BCC) > 0 {
		msg.SetHeader("Bcc", letter.RecipientsHeader(let.BCC))
	}

	if let.Subject != "" {
		msg.SetHeader("Subject", let.Subject)
	}

	if let.Text != "" {
		msg.SetBody("text/plain", let.Text)
	}

	if let.HTML != "" {
		msg.SetBody("text/html", let.HTML)
	}

	for _, a := range let.Attachments {
		msg.AttachReader(a.Filename, bytes.NewReader(a.Content), mail.SetHeader(map[string][]string{
			"Content-Type":              {a.Header.Get("Content-Type")},
			"Content-Disposition":       {a.Header.Get("Content-Disposition")},
			"Content-ID":                {a.Header.Get("Content-ID")},
			"Content-Transfer-Encoding": {a.Header.Get("Content-Transfer-Encoding")},
		}))
	}

	return trans.dialer.DialAndSend(msg)
}
