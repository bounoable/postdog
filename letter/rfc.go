package letter

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
)

type rfcMessage struct {
	from        string
	to          string
	cc          string
	bcc         string
	subject     string
	text        string
	html        string
	attachments []Attachment
	lines       []string
}

func rfc(from, to, cc, bcc, subject, text, html string, attachments []Attachment) *rfcMessage {
	msg := &rfcMessage{
		from:        from,
		to:          to,
		cc:          cc,
		bcc:         bcc,
		subject:     subject,
		text:        text,
		html:        html,
		attachments: attachments,
	}

	msg.build()

	return msg
}

func (msg *rfcMessage) line(lines ...string) {
	msg.lines = append(msg.lines, lines...)
}

func (msg *rfcMessage) keyValue(key, val string) {
	msg.line(fmt.Sprintf("%s: %s", key, val))
}

func (msg *rfcMessage) contentType(ct string, fn func(msg *rfcMessage, bd string)) {
	bd := boundary()
	msg.keyValue("Content-Type", fmt.Sprintf("%s; boundary=%s", ct, bd))
	fn(msg, bd)
}

func (msg *rfcMessage) contentTypeIf(cond bool, ct string, fn func(msg *rfcMessage, bd string)) {
	if cond {
		msg.contentType(ct, fn)
		return
	}

	fn(msg, "")
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func boundary() string {
	buf := make([]byte, 64)
	for i := 0; i < 64; i++ {
		buf[i] = chars[rand.Intn(62)]
	}
	return string(buf)
}

func (msg *rfcMessage) beginBoundary(bd string) {
	if bd == "" {
		return
	}

	msg.line(fmt.Sprintf("--%s", bd))
}

func (msg *rfcMessage) endBoundary(bd string) {
	if bd == "" {
		return
	}

	msg.line(fmt.Sprintf("--%s--", bd))
}

func (msg *rfcMessage) build() {
	msg.lines = nil

	msg.keyValue("MIME-Version", "1.0")
	msg.keyValue("Subject", encodeSubject(msg.subject))
	msg.keyValue("From", msg.from)

	if msg.to != "" {
		msg.keyValue("To", msg.to)
	}

	if msg.cc != "" {
		msg.keyValue("Cc", msg.cc)
	}

	if msg.bcc != "" {
		msg.keyValue("Bcc", msg.bcc)
	}

	msg.contentTypeIf(len(msg.attachments) > 0, "multipart/mixed", func(msg *rfcMessage, bd string) {
		endMixedBoundary := true

		msg.beginBoundary(bd)
		msg.contentTypeIf(msg.text != "" && msg.html != "", "multipart/alternative", func(msg *rfcMessage, bd string) {
			if strings.TrimSpace(msg.text) != "" {
				msg.beginBoundary(bd)
				msg.keyValue("Content-Type", `text/plain; charset="utf-8"`)
				msg.keyValue("Content-Transfer-Encoding", "base64")
				msg.line("", base64.StdEncoding.EncodeToString([]byte(msg.text)))
			}

			if strings.TrimSpace(msg.html) != "" {
				msg.beginBoundary(bd)
				msg.keyValue("Content-Type", `text/html; charset="utf-8"`)
				msg.keyValue("Content-Transfer-Encoding", "base64")
				msg.line("", base64.StdEncoding.EncodeToString([]byte(msg.html)))
			}

			if msg.text != "" || msg.html != "" {
				msg.endBoundary(bd)
				return
			}

			endMixedBoundary = false
		})

		if len(msg.attachments) > 0 {
			endMixedBoundary = true

			for _, a := range msg.attachments {
				msg.beginBoundary(bd)
				msg.keyValue("Content-Type", a.Header.Get("Content-Type"))
				msg.keyValue("Content-Disposition", a.Header.Get("Content-Disposition"))
				msg.keyValue("Content-ID", a.Header.Get("Content-ID"))
				msg.keyValue("Content-Transfer-Encoding", a.Header.Get("Content-Transfer-Encoding"))
				msg.line("", base64.StdEncoding.EncodeToString(a.Content))
			}
		}

		if endMixedBoundary {
			msg.endBoundary(bd)
		}
	})
}

func (msg *rfcMessage) String() string {
	return strings.Join(msg.lines, "\r\n")
}

func encodeSubject(subject string) string {
	return fmt.Sprintf("=?utf-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
}
