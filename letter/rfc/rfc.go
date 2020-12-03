package rfc

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/bounoable/postdog/internal/encode"
)

// Build the mail according to RFC 5322.
func Build(mail Mail, opts ...Option) string {
	var b builder
	for _, opt := range opts {
		opt(&b)
	}
	if b.newBoundary == nil {
		b.newBoundary = newBoundary
	}
	return b.build(mail)
}

// Mail contains the data of a mail.
type Mail struct {
	Subject     string
	From        mail.Address
	To          []mail.Address
	CC          []mail.Address
	BCC         []mail.Address
	ReplyTo     []mail.Address
	Text        string
	HTML        string
	Attachments []Attachment
}

// Attachment is a mail attachment.
type Attachment struct {
	Filename string
	Content  []byte
	Header   textproto.MIMEHeader
}

// Option is a builder option.
type Option func(*builder)

type builder struct {
	newBoundary func() string
}

// DeterministicBoundaries forces the builder to generate the encapsulation boundaries deterministically.
func DeterministicBoundaries() func(b *builder) {
	var index int
	return func(b *builder) {
		b.newBoundary = func() string {
			bd := fmt.Sprintf("%064d", index+1)
			index++
			return bd
		}
	}
}

var emptyAddr mail.Address

func (b builder) build(mail Mail) string {
	lines := []string{"MIME-Version: 1.0"}

	if mail.Subject != "" {
		lines = append(lines, fmt.Sprintf("Subject: %s", encode.UTF8(mail.Subject)))
	}

	if mail.From != emptyAddr {
		lines = append(lines, fmt.Sprintf("From: %s", mail.From.String()))
	}

	if len(mail.To) > 0 {
		lines = append(lines, fmt.Sprintf("To: %s", joinAddresses(mail.To...)))
	}

	if len(mail.CC) > 0 {
		lines = append(lines, fmt.Sprintf("Cc: %s", joinAddresses(mail.CC...)))
	}

	if len(mail.BCC) > 0 {
		lines = append(lines, fmt.Sprintf("Bcc: %s", joinAddresses(mail.BCC...)))
	}

	if len(mail.ReplyTo) > 0 {
		lines = append(lines, fmt.Sprintf("Reply-To: %s", joinAddresses(mail.ReplyTo...)))
	}

	textLines := b.textLines(mail.Text)
	htmlLines := b.htmlLines(mail.HTML)

	if len(mail.Attachments) == 0 {
		return strings.Join(append(lines, b.bodyWithoutAttachments(textLines, htmlLines)...), "\r\n")
	}

	lines = append(lines, b.contentType("multipart/mixed", func(bd string) []string {
		lines := append([]string{startBoundary(bd)}, b.bodyWithoutAttachments(textLines, htmlLines)...)
		for _, at := range mail.Attachments {
			lines = append(
				lines,
				startBoundary(bd),
				fmt.Sprintf("Content-Type: %s", at.Header.Get("Content-Type")),
				fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len(at.Content), encode.UTF8(at.Filename)),
				fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum(at.Content))[:12], encode.ToASCII(at.Filename)),
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString(at.Content), 78),
				"",
			)
		}
		return append(lines, endBoundary(bd))
	})...)

	return strings.Join(lines, "\r\n")
}

func (b builder) bodyWithoutAttachments(text []string, html []string) (lines []string) {
	if len(text) > 0 && len(html) > 0 {
		lines = append(lines, b.contentType("multipart/alternative", func(bd string) []string {
			lines := append([]string{startBoundary(bd)}, text...)
			lines = append(lines, append([]string{startBoundary(bd)}, html...)...)
			lines = append(lines, endBoundary(bd))
			return lines
		})...)
	} else if len(text) > 0 {
		lines = append(lines, text...)
	} else if len(html) > 0 {
		lines = append(lines, html...)
	}
	return
}

func (b builder) contentType(ct string, fn func(string) []string) []string {
	bd := b.newBoundary()
	lines := []string{fmt.Sprintf(`Content-Type: %s; boundary="%s"`, ct, bd), "", ""}
	return append(lines, fn(bd)...)
}

func (b builder) textLines(text string) []string {
	if text == "" {
		return nil
	}

	return []string{
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: base64",
		"",
		fold(base64.StdEncoding.EncodeToString([]byte(text)), 78),
		"",
	}
}

func (b builder) htmlLines(html string) []string {
	if html == "" {
		return nil
	}

	return []string{
		"Content-Type: text/html; charset=utf-8",
		"Content-Transfer-Encoding: base64",
		"",
		fold(base64.StdEncoding.EncodeToString([]byte(html)), 78),
		"",
	}
}

func joinAddresses(addrs ...mail.Address) string {
	addrstrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrstrs[i] = addr.String()
	}
	return strings.Join(addrstrs, ",")
}

func startBoundary(bd string) string {
	return fmt.Sprintf("--%s", bd)
}

func endBoundary(bd string) string {
	return fmt.Sprintf("%s--", startBoundary(bd))
}

func fold(s string, after int) string {
	sub := ""
	subs := []string{}
	runes := []rune(s)
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i+1)%after == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}
	return strings.Join(subs, "\r\n")
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func newBoundary() string {
	buf := make([]byte, 64)
	for i := 0; i < 64; i++ {
		buf[i] = chars[rand.Intn(62)]
	}
	return string(buf)
}
