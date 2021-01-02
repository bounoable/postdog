package letter_test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/bounoable/postdog/internal/encode"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/letter/rfc"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		name          string
		opts          []letter.Option
		expect        func(*testing.T, letter.Letter)
		expectedError error
	}{
		{
			name: "Subject()",
			opts: []letter.Option{
				letter.Subject("Hi."),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, "Hi.", l.Subject())
			},
		},
		{
			name: "From()",
			opts: []letter.Option{
				letter.From("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}, l.From())
			},
		},
		{
			name: "Recipient()",
			opts: []letter.Option{
				letter.Recipient("Bob Belcher", "bob@example.com"),
				letter.Recipient("Linda Belcher", "linda@example.com"),
				letter.To("Tina Belcher", "tina@example.com"),
				letter.CC("Gene Belcher", "gene@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
					{Name: "Tina Belcher", Address: "tina@example.com"},
					{Name: "Gene Belcher", Address: "gene@example.com"},
				}, l.Recipients())
			},
		},
		{
			name: "To(): single recipient",
			opts: []letter.Option{
				letter.To("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.To())
			},
		},
		{
			name: "To(): multiple recipients",
			opts: []letter.Option{
				letter.To("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				}, l.To())
			},
		},
		{
			name: "To(): dedupe",
			opts: []letter.Option{
				letter.To("Bob Belcher", "bob@example.com"),
				letter.To("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.To())
			},
		},
		{
			name: "CC(): single recipient",
			opts: []letter.Option{
				letter.CC("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.CC())
			},
		},
		{
			name: "CC(): multiple recipients",
			opts: []letter.Option{
				letter.CC("Bob Belcher", "bob@example.com"),
				letter.CC("Linda Belcher", "linda@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				}, l.CC())
			},
		},
		{
			name: "CC(): dedupe",
			opts: []letter.Option{
				letter.CC("Bob Belcher", "bob@example.com"),
				letter.CC("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.CC())
			},
		},
		{
			name: "BCC(): single recipient",
			opts: []letter.Option{
				letter.BCC("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.BCC())
			},
		},
		{
			name: "BCC(): multiple recipients",
			opts: []letter.Option{
				letter.BCC("Bob Belcher", "bob@example.com"),
				letter.BCC("Linda Belcher", "linda@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				}, l.BCC())
			},
		},
		{
			name: "BCC(): dedupe",
			opts: []letter.Option{
				letter.BCC("Bob Belcher", "bob@example.com"),
				letter.BCC("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.BCC())
			},
		},
		{
			name: "ReplyTo(): single recipient",
			opts: []letter.Option{
				letter.ReplyTo("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.ReplyTo())
			},
		},
		{
			name: "ReplyTo(): multiple recipients",
			opts: []letter.Option{
				letter.ReplyTo("Bob Belcher", "bob@example.com"),
				letter.ReplyTo("Linda Belcher", "linda@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
					{Name: "Linda Belcher", Address: "linda@example.com"},
				}, l.ReplyTo())
			},
		},
		{
			name: "ReplyTo(): dedupe",
			opts: []letter.Option{
				letter.ReplyTo("Bob Belcher", "bob@example.com"),
				letter.ReplyTo("Bob Belcher", "bob@example.com"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, []mail.Address{
					{Name: "Bob Belcher", Address: "bob@example.com"},
				}, l.ReplyTo())
			},
		},
		{
			name: "Text()",
			opts: []letter.Option{
				letter.Text("Hello."),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, "Hello.", l.Text())
			},
		},
		{
			name: "HTML()",
			opts: []letter.Option{
				letter.HTML("<p>Hello.</p>"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, "<p>Hello.</p>", l.HTML())
			},
		},
		{
			name: "Content()",
			opts: []letter.Option{
				letter.Content("Hello.", "<p>Hello.</p>"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, "Hello.", l.Text())
				assert.Equal(t, "<p>Hello.</p>", l.HTML())
				text, html := l.Content()
				assert.Equal(t, "Hello.", text)
				assert.Equal(t, "<p>Hello.</p>", html)
			},
		},
		{
			name: "RFC()",
			opts: []letter.Option{
				letter.RFC("rfc body"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				assert.Equal(t, "rfc body", l.RFC())
			},
		},
		{
			name: "Attach(): detect content-type by content",
			opts: []letter.Option{
				letter.Attach("attach1", []byte{1, 2, 3}),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, 3, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assert.Equal(t, "application/octet-stream", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "Attach(): detect content-type by extension",
			opts: []letter.Option{
				letter.Attach("attach1.html", []byte{1, 2, 3}),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1.html", at.Filename())
				assert.Equal(t, 3, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assert.Equal(t, "text/html; charset=utf-8", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "Attach(): explicit content-type",
			opts: []letter.Option{
				letter.Attach("attach1", []byte{1, 2, 3}, letter.AttachmentType("application/json")),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, 3, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assert.Equal(t, "application/json", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "Attach(): explicit content length",
			opts: []letter.Option{
				letter.Attach("attach1", []byte{1, 2, 3}, letter.AttachmentSize(5)),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, 5, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "Attach(): multiple attachments",
			opts: []letter.Option{
				letter.Attach("attach1", []byte{1, 2, 3}),
				letter.Attach("attach2", []byte{2, 3, 4, 5}, letter.AttachmentType("application/json")),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, 3, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assert.Equal(t, "application/octet-stream", at.ContentType())
				assertAttachmentHeader(t, at)

				at = l.Attachments()[1]
				assert.Equal(t, "attach2", at.Filename())
				assert.Equal(t, 4, at.Size())
				assert.Equal(t, []byte{2, 3, 4, 5}, at.Content())
				assert.Equal(t, "application/json", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "AttachReader()",
			opts: []letter.Option{
				letter.AttachReader("attach1", bytes.NewReader([]byte{1, 2, 3})),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, 3, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assert.Equal(t, "application/octet-stream", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "AttachReader(): error",
			opts: []letter.Option{
				letter.AttachReader("attach1", mockReader{readFn: func([]byte) (int, error) {
					return 0, mockError
				}}),
			},
			expectedError: mockError,
		},
		{
			name: "AttachFile()",
			opts: []letter.Option{
				letter.AttachFile("attach1", "./testdata/attachment.txt"),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, "Hello.\n", string(at.Content()))
				assert.Equal(t, "text/plain; charset=utf-8", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "AttachFile(): explicit content-type",
			opts: []letter.Option{
				letter.AttachFile("attach1", "./testdata/attachment.txt", letter.AttachmentType("text/html")),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1", at.Filename())
				assert.Equal(t, "Hello.\n", string(at.Content()))
				assert.Equal(t, "text/html", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			let, err := letter.TryWrite(test.opts...)
			assert.True(t, errors.Is(err, test.expectedError))
			if test.expect != nil {
				test.expect(t, let)
			}
		})
	}
}

func assertAttachmentHeader(t *testing.T, at letter.Attachment) {
	assert.Equal(t, fmt.Sprintf(`%s; name="%s"`, at.ContentType(), encode.UTF8(at.Filename())), at.Header().Get("Content-Type"))
	assert.Equal(t, fmt.Sprintf(`attachment; size=%d; filename="%s"`, at.Size(), encode.UTF8(at.Filename())), at.Header().Get("Content-Disposition"))
	assert.Equal(t, fmt.Sprintf("<%s_%s>", fmt.Sprintf("%x", sha1.Sum(at.Content()))[:12], encode.ToASCII(at.Filename())), at.Header().Get("Content-ID"))
	assert.Equal(t, "base64", at.Header().Get("Content-Transfer-Encoding"))
}

type mockReader struct {
	readFn func([]byte) (int, error)
}

func (mr mockReader) Read(b []byte) (int, error) {
	return mr.readFn(b)
}

var mockError = errors.New("mock error")

func TestLetter_Recipients(t *testing.T) {
	tests := []struct {
		name     string
		opts     []letter.Option
		expected []mail.Address
	}{
		{
			name: "no recipients",
		},
		{
			name: "To() recipients",
			opts: []letter.Option{
				letter.To("Bob Belcher", "bob@example.com"),
				letter.To("Linda Belcher", "linda@example.com"),
			},
			expected: []mail.Address{
				{Name: "Bob Belcher", Address: "bob@example.com"},
				{Name: "Linda Belcher", Address: "linda@example.com"},
			},
		},
		{
			name: "CC() recipients",
			opts: []letter.Option{
				letter.CC("Bob Belcher", "bob@example.com"),
				letter.CC("Linda Belcher", "linda@example.com"),
			},
			expected: []mail.Address{
				{Name: "Bob Belcher", Address: "bob@example.com"},
				{Name: "Linda Belcher", Address: "linda@example.com"},
			},
		},
		{
			name: "BCC() recipients",
			opts: []letter.Option{
				letter.BCC("Bob Belcher", "bob@example.com"),
				letter.BCC("Linda Belcher", "linda@example.com"),
			},
			expected: []mail.Address{
				{Name: "Bob Belcher", Address: "bob@example.com"},
				{Name: "Linda Belcher", Address: "linda@example.com"},
			},
		},
		{
			name: "mixed recipients",
			opts: []letter.Option{
				letter.To("Bob Belcher", "bob@example.com"),
				letter.CC("Linda Belcher", "linda@example.com"),
				letter.BCC("Tina Belcher", "tina@example.com"),
			},
			expected: []mail.Address{
				{Name: "Bob Belcher", Address: "bob@example.com"},
				{Name: "Linda Belcher", Address: "linda@example.com"},
				{Name: "Tina Belcher", Address: "tina@example.com"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			let, err := letter.TryWrite(test.opts...)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, let.Recipients())
		})
	}
}

func TestLetter_WithSubject(t *testing.T) {
	assert.Equal(t, "foo", letter.Write().WithSubject("foo").Subject())
}

func TestLetter_WithFrom(t *testing.T) {
	addr := mail.Address{
		Name:    "Bob",
		Address: "bob@example.com",
	}
	assert.Equal(t, addr, letter.Write().WithFrom(addr.Name, addr.Address).From())
	assert.Equal(t, addr, letter.Write().WithFromAddress(addr).From())
}

func TestLetter_WithRecipients(t *testing.T) {
	addrs := []mail.Address{
		{
			Name:    "Bob",
			Address: "bob@example.com",
		},
		{
			Name:    "Linda",
			Address: "linda@example.com",
		},
	}
	assert.Equal(t, addrs, letter.Write().WithRecipients(addrs...).Recipients())
}

func TestLetter_WithTo(t *testing.T) {
	addrs := []mail.Address{
		{
			Name:    "Bob",
			Address: "bob@example.com",
		},
		{
			Name:    "Linda",
			Address: "linda@example.com",
		},
	}
	assert.Equal(t, addrs, letter.Write().WithTo(addrs...).To())
}

func TestLetter_WithCC(t *testing.T) {
	addrs := []mail.Address{
		{
			Name:    "Bob",
			Address: "bob@example.com",
		},
		{
			Name:    "Linda",
			Address: "linda@example.com",
		},
	}
	assert.Equal(t, addrs, letter.Write().WithCC(addrs...).CC())
}
func TestLetter_WithBCC(t *testing.T) {
	addrs := []mail.Address{
		{
			Name:    "Bob",
			Address: "bob@example.com",
		},
		{
			Name:    "Linda",
			Address: "linda@example.com",
		},
	}
	assert.Equal(t, addrs, letter.Write().WithBCC(addrs...).BCC())
}

func TestLetter_WithReplyTo(t *testing.T) {
	addrs := []mail.Address{
		{
			Name:    "Bob",
			Address: "bob@example.com",
		},
		{
			Name:    "Linda",
			Address: "linda@example.com",
		},
	}
	assert.Equal(t, addrs, letter.Write().WithReplyTo(addrs...).ReplyTo())
}

func TestLetter_WithText(t *testing.T) {
	assert.Equal(t, "foo", letter.Write().WithText("foo").Text())
}

func TestLetter_WithHTML(t *testing.T) {
	assert.Equal(t, "foo", letter.Write().WithHTML("foo").HTML())
}

func TestLetter_WithContent(t *testing.T) {
	assert.Equal(t, letter.Write(letter.Text("foo"), letter.HTML("bar")), letter.Write().WithContent("foo", "bar"))
}

func TestLetter_WithAttachments(t *testing.T) {
	give := letter.Write()
	attach := []letter.Option{
		letter.Attach("foo", []byte{1, 2, 3}),
		letter.Attach("bar", []byte{1, 2, 3}),
	}
	want := letter.Write(attach...)
	assert.Equal(t, want, give.WithAttachments(
		letter.NewAttachment("foo", []byte{1, 2, 3}),
		letter.NewAttachment("bar", []byte{1, 2, 3}),
	))
}

func TestLetter_WithRFC(t *testing.T) {
	assert.Equal(t, letter.Write(letter.RFC("rfc body")), letter.Write().WithRFC("rfc body"))
}

func TestLetter_RFC(t *testing.T) {
	clock := staticClock(time.Now())

	let, err := letter.TryWrite(
		letter.Subject("Hi."),
		letter.Text("Hello."),
		letter.HTML("<p>Hello.</p>"),
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
		letter.CC("Gene Belcher", "gene@example.com"),
		letter.CC("Tina Belcher", "tina@example.com"),
		letter.BCC("Jimmy Pesto", "jimmy@example.com"),
		letter.BCC("Jimmy Pesto Jr.", "jimmyjr@example.com"),
		letter.ReplyTo("Bosco", "bosco@example.com"),
		letter.ReplyTo("Teddy", "teddy@example.com"),
		letter.Attach("attach1", []byte("Attachment 1"), letter.AttachmentType("text/plain")),
		letter.Attach("attach2", []byte("<p>Attachment 2</p>"), letter.AttachmentType("text/html")),
	)
	assert.Nil(t, err)

	expected := join(
		"MIME-Version: 1.0",
		"Message-ID: foobar",
		fmt.Sprintf("Date: %s", clock.Now().Format(time.RFC1123Z)),
		fmt.Sprintf("Subject: %s", encode.UTF8("Hi.")),
		`From: "Bob Belcher" <bob@example.com>`,
		`To: "Linda Belcher" <linda@example.com>`,
		`Cc: "Gene Belcher" <gene@example.com>,"Tina Belcher" <tina@example.com>`,
		`Bcc: "Jimmy Pesto" <jimmy@example.com>,"Jimmy Pesto Jr." <jimmyjr@example.com>`,
		`Reply-To: "Bosco" <bosco@example.com>,"Teddy" <teddy@example.com>`,

		fmt.Sprintf(`Content-Type: multipart/mixed; boundary="%s"`, boundary(0)),
		"", "", // preamble

		startBoundary(0),
		fmt.Sprintf(`Content-Type: multipart/alternative; boundary="%s"`, boundary(1)),
		"", "", // preamble

		startBoundary(1),
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: base64",
		"",
		fold(base64.StdEncoding.EncodeToString([]byte("Hello.")), 78),
		"",

		startBoundary(1),
		"Content-Type: text/html; charset=utf-8",
		"Content-Transfer-Encoding: base64",
		"",
		fold(base64.StdEncoding.EncodeToString([]byte("<p>Hello.</p>")), 78),
		"",

		endBoundary(1),

		startBoundary(0),
		fmt.Sprintf(`Content-Type: text/plain; name="%s"`, encode.UTF8("attach1")),
		fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len([]byte("Attachment 1")), encode.UTF8("attach1")),
		fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum([]byte("Attachment 1")))[:12], encode.ToASCII("attach1")),
		fmt.Sprintf("Content-Transfer-Encoding: base64"),
		"",
		fold(base64.StdEncoding.EncodeToString([]byte("Attachment 1")), 78),
		"",

		startBoundary(0),
		fmt.Sprintf(`Content-Type: text/html; name="%s"`, encode.UTF8("attach2")),
		fmt.Sprintf(`Content-Disposition: attachment; size=%d; filename="%s"`, len([]byte("<p>Attachment 2</p>")), encode.UTF8("attach2")),
		fmt.Sprintf("Content-ID: <%s_%s>", fmt.Sprintf("%x", sha1.Sum([]byte("<p>Attachment 2</p>")))[:12], encode.ToASCII("attach2")),
		fmt.Sprintf("Content-Transfer-Encoding: base64"),
		"",
		fold(base64.StdEncoding.EncodeToString([]byte("<p>Attachment 2</p>")), 78),
		"",

		endBoundary(0),
	)

	assert.Equal(t, expected, let.WithRFCOptions(rfc.WithClock(clock), rfc.WithMessageID("foobar")).RFC())
}

func TestLetter_RFC_override(t *testing.T) {
	let := letter.Write(
		letter.From("Bob Belcher", "bob@example.com"),
		letter.To("Linda Belcher", "linda@example.com"),
		letter.Subject("Hi."),
		letter.Text("Hello."),
		letter.HTML("<p>Hello.</p>"),
		letter.RFC("rfc body"),
	)

	assert.Equal(t, "rfc body", let.RFC())
}

func join(lines ...string) string {
	return strings.Join(lines, "\r\n")
}

func boundary(i int) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%064d", i+1))))
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

func staticClock(t time.Time) rfc.Clock {
	return rfc.ClockFunc(func() time.Time { return t })
}
