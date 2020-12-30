package rfc_test

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bounoable/postdog/internal/encode"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/letter/rfc"
	"github.com/stretchr/testify/assert"
)

var (
	baseLetterOpts = []letter.Option{
		letter.Subject("Hi."),
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
	}
)

func TestBuild(t *testing.T) {
	clock := staticClock(time.Now())
	idgen := staticID("<id@domain>")

	tests := []struct {
		name       string
		letterOpts []letter.Option
		expected   string
	}{
		{
			name: "basic",
			letterOpts: append(baseLetterOpts,
				letter.Text("Hello."),
				letter.HTML("<p>Hello.</p>"),
			),
			expected: join(
				"MIME-Version: 1.0",
				"Message-ID: <id@domain>",
				fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
				fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
				`From: "Bob Belcher" <bob@example.com>`,
				`To: "Linda Belcher" <linda@example.com>`,
				fmt.Sprintf(`Content-Type: multipart/alternative; boundary="%s"`, boundary(0)),
				"",
				"",
				startBoundary(0),
				"Content-Type: text/plain; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("Hello.")), 76),
				"",
				startBoundary(0),
				"Content-Type: text/html; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("<p>Hello.</p>")), 76),
				"",
				endBoundary(0),
			),
		},
		{
			name:       "basic, no html",
			letterOpts: append(baseLetterOpts, letter.Text("Hello.")),
			expected: join(
				"MIME-Version: 1.0",
				"Message-ID: <id@domain>",
				fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
				fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
				`From: "Bob Belcher" <bob@example.com>`,
				`To: "Linda Belcher" <linda@example.com>`,
				"Content-Type: text/plain; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("Hello.")), 76),
				"",
			),
		},
		{
			name:       "basic, no text",
			letterOpts: append(baseLetterOpts, letter.HTML("<p>Hello.</p>")),
			expected: join(
				"MIME-Version: 1.0",
				"Message-ID: <id@domain>",
				fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
				fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
				`From: "Bob Belcher" <bob@example.com>`,
				`To: "Linda Belcher" <linda@example.com>`,
				"Content-Type: text/html; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("<p>Hello.</p>")), 76),
				"",
			),
		},
		{
			name: "optional recipients",
			letterOpts: append(
				baseLetterOpts,
				letter.CC("Gene Belcher", "gene@example.com"),
				letter.CC("Tina Belcher", "tina@example.com"),
				letter.BCC("Jimmy Pesto", "jimmy@example.com"),
				letter.BCC("Jimmy Pesto Jr.", "jimmyjr@example.com"),
				letter.ReplyTo("Bosco", "bosco@example.com"),
				letter.ReplyTo("Teddy", "teddy@example.com"),
			),
			expected: join(
				"MIME-Version: 1.0",
				"Message-ID: <id@domain>",
				fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
				fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
				`From: "Bob Belcher" <bob@example.com>`,
				`To: "Linda Belcher" <linda@example.com>`,
				`Cc: "Gene Belcher" <gene@example.com>,"Tina Belcher" <tina@example.com>`,
				`Bcc: "Jimmy Pesto" <jimmy@example.com>,"Jimmy Pesto Jr." <jimmyjr@example.com>`,
				`Reply-To: "Bosco" <bosco@example.com>,"Teddy" <teddy@example.com>`,
			),
		},
		{
			name: "text & html with attachments",
			letterOpts: append(
				baseLetterOpts,
				letter.Text("Hello."),
				letter.HTML("<p>Hello.</p>"),
				letter.Attach("attach1", []byte("Attachment 1"), letter.AttachmentType("text/plain")),
				letter.Attach("attach2", []byte("<p>Attachment 2</p>"), letter.AttachmentType("text/html")),
			),
			expected: join(
				"MIME-Version: 1.0",
				"Message-ID: <id@domain>",
				fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
				fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
				`From: "Bob Belcher" <bob@example.com>`,
				`To: "Linda Belcher" <linda@example.com>`,

				fmt.Sprintf(`Content-Type: multipart/mixed; boundary="%s"`, boundary(0)),
				"", "", // preamble

				startBoundary(0),
				fmt.Sprintf(`Content-Type: multipart/alternative; boundary="%s"`, boundary(1)),
				"", "", // preamble

				startBoundary(1),
				"Content-Type: text/plain; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("Hello.")), 76),
				"",

				startBoundary(1),
				"Content-Type: text/html; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("<p>Hello.</p>")), 76),
				"",

				endBoundary(1),

				startBoundary(0),
				fmt.Sprintf(`Content-Type: text/plain; name="%s"`, encode.UTF8("attach1")),
				fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len([]byte("Attachment 1")), encode.UTF8("attach1")),
				fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum([]byte("Attachment 1")))[:12], encode.ToASCII("attach1")),
				fmt.Sprintf("Content-Transfer-Encoding: base64"),
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("Attachment 1")), 76),
				"",

				startBoundary(0),
				fmt.Sprintf(`Content-Type: text/html; name="%s"`, encode.UTF8("attach2")),
				fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len([]byte("<p>Attachment 2</p>")), encode.UTF8("attach2")),
				fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum([]byte("<p>Attachment 2</p>")))[:12], encode.ToASCII("attach2")),
				fmt.Sprintf("Content-Transfer-Encoding: base64"),
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("<p>Attachment 2</p>")), 76),
				"",

				endBoundary(0),
			),
		},
		{
			name: "html with attachments",
			letterOpts: append(
				baseLetterOpts,
				letter.HTML("<p>Hello.</p>"),
				letter.Attach("attach1", []byte("Attachment 1"), letter.AttachmentType("text/plain")),
				letter.Attach("attach2", []byte("<p>Attachment 2</p>"), letter.AttachmentType("text/html")),
			),
			expected: join(
				"MIME-Version: 1.0",
				"Message-ID: <id@domain>",
				fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
				fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
				`From: "Bob Belcher" <bob@example.com>`,
				`To: "Linda Belcher" <linda@example.com>`,

				fmt.Sprintf(`Content-Type: multipart/mixed; boundary="%s"`, boundary(0)),
				"", "", // preamble

				startBoundary(0),
				"Content-Type: text/html; charset=utf-8",
				"Content-Transfer-Encoding: base64",
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("<p>Hello.</p>")), 76),
				"",

				startBoundary(0),
				fmt.Sprintf(`Content-Type: text/plain; name="%s"`, encode.UTF8("attach1")),
				fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len([]byte("Attachment 1")), encode.UTF8("attach1")),
				fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum([]byte("Attachment 1")))[:12], encode.ToASCII("attach1")),
				fmt.Sprintf("Content-Transfer-Encoding: base64"),
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("Attachment 1")), 76),
				"",

				startBoundary(0),
				fmt.Sprintf(`Content-Type: text/html; name="%s"`, encode.UTF8("attach2")),
				fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len([]byte("<p>Attachment 2</p>")), encode.UTF8("attach2")),
				fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum([]byte("<p>Attachment 2</p>")))[:12], encode.ToASCII("attach2")),
				fmt.Sprintf("Content-Transfer-Encoding: base64"),
				"",
				fold(base64.StdEncoding.EncodeToString([]byte("<p>Attachment 2</p>")), 76),
				"",

				endBoundary(0),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			let, err := letter.TryWrite(test.letterOpts...)
			assert.Nil(t, err)

			s := rfc.Build(rfc.Mail{
				Subject:     let.Subject(),
				From:        let.From(),
				To:          let.To(),
				CC:          let.CC(),
				BCC:         let.BCC(),
				ReplyTo:     let.ReplyTo(),
				Text:        let.Text(),
				HTML:        let.HTML(),
				Attachments: mapAttachments(let.Attachments()...),
			}, rfc.WithClock(clock), rfc.WithIDGenerator(idgen))

			assert.Equal(t, test.expected, s)
		})
	}
}

func join(lines ...string) string {
	return strings.Join(lines, "\r\n")
}

func boundary(i int) string {
	v := fmt.Sprintf("%064d", i+1)
	s := md5.Sum([]byte(v))
	return fmt.Sprintf("%x", s)
}

func startBoundary(i int) string {
	return fmt.Sprintf("--%s", boundary(i))
}

func endBoundary(i int) string {
	return fmt.Sprintf("%s--", startBoundary(i))
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

func mapAttachments(ats ...letter.Attachment) []rfc.Attachment {
	res := make([]rfc.Attachment, len(ats))
	for i, at := range ats {
		res[i] = rfc.Attachment{
			Filename: at.Filename(),
			Content:  at.Content(),
			Header:   at.Header(),
		}
	}
	return res
}

func staticClock(t time.Time) rfc.Clock {
	return rfc.ClockFunc(func() time.Time {
		return t
	})
}

func staticID(id string) rfc.MessageIDFactory {
	return rfc.MessageIDFunc(func(rfc.Mail) string { return id })
}
