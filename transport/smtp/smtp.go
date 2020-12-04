package smtp

//go:generate mockgen -source=smtp.go -destination=./mocks/smtp.go

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/bounoable/postdog"
)

// Transport returns an SMTP transport.
func Transport(host string, port int, username, password string) postdog.Transport {
	return TransportWithSender(smtpSender{}, host, port, username, password)
}

// TransportWithSender returns an SMTP transport and accepts a custom implementation of the smtp.SendMail() function.
func TransportWithSender(sender MailSender, host string, port int, username, password string) postdog.Transport {
	return &transport{
		sender:   sender,
		host:     host,
		port:     port,
		username: username,
		password: password,
		addr:     fmt.Sprintf("%s:%d", host, port),
		auth:     smtp.PlainAuth("", username, password, host),
	}
}

// MailSender wraps the smtp.SendMail() function in an interface.
type MailSender interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

type transport struct {
	sender   MailSender
	host     string
	port     int
	username string
	password string

	addr string
	auth smtp.Auth
}

func (tr *transport) Send(_ context.Context, m postdog.Mail) error {
	to := make([]string, len(m.Recipients()))
	for i, rcpt := range m.Recipients() {
		to[i] = rcpt.Address
	}
	return tr.sender.SendMail(tr.addr, tr.auth, m.From().Address, to, []byte(m.RFC()))
}

type smtpSender struct{}

func (s smtpSender) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, msg)
}
