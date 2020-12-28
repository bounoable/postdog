package smtp

//go:generate mockgen -source=smtp.go -destination=./mocks/smtp.go

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter/rfc"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

// MailSender wraps the smtp.SendMail() function in an interface.
type MailSender interface {
	SendMail(addr string, a sasl.Client, from string, to []string, msg []byte) error
}

type transport struct {
	sender   MailSender
	host     string
	port     int
	username string
	password string

	addr string
	auth sasl.Client

	rfcOpts []rfc.Option
}

type smtpSender struct{}

// Option is a transport option.
type Option func(*transport)

// RFCOptions returns an Option that adds rfc.Options when generating the RFC body.
func RFCOptions(opts ...rfc.Option) Option {
	return func(t *transport) {
		t.rfcOpts = append(t.rfcOpts, opts...)
	}
}

// Transport returns an SMTP transport.
func Transport(host string, port int, username, password string, opts ...Option) postdog.Transport {
	return TransportWithSender(smtpSender{}, host, port, username, password, opts...)
}

// TransportWithSender returns an SMTP transport and accepts a custom implementation of the smtp.SendMail() function.
func TransportWithSender(sender MailSender, host string, port int, username, password string, opts ...Option) postdog.Transport {
	t := &transport{
		sender:   sender,
		host:     host,
		port:     port,
		username: username,
		password: password,
		addr:     fmt.Sprintf("%s:%d", host, port),
		auth:     sasl.NewPlainClient("", username, password),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (tr *transport) Send(_ context.Context, m postdog.Mail) error {
	to := make([]string, len(m.Recipients()))
	for i, rcpt := range m.Recipients() {
		to[i] = rcpt.Address
	}
	return tr.sender.SendMail(tr.addr, tr.auth, m.From().Address, to, []byte(m.RFC(tr.rfcOpts...)))
}

func (s smtpSender) SendMail(addr string, a sasl.Client, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, bytes.NewReader(msg))
}
