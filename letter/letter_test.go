package letter_test

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"testing"

	"github.com/bounoable/postdog/internal/encode"
	"github.com/bounoable/postdog/letter"
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
				letter.Attach("attach1.txt", []byte{1, 2, 3}),
			},
			expect: func(t *testing.T, l letter.Letter) {
				at := l.Attachments()[0]
				assert.Equal(t, "attach1.txt", at.Filename())
				assert.Equal(t, 3, at.Size())
				assert.Equal(t, []byte{1, 2, 3}, at.Content())
				assert.Equal(t, "text/plain; charset=utf-8", at.ContentType())
				assertAttachmentHeader(t, at)
			},
		},
		{
			name: "Attach(): explicit content-type",
			opts: []letter.Option{
				letter.Attach("attach1", []byte{1, 2, 3}, letter.ContentType("application/json")),
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
			name: "Attach(): multiple attachments",
			opts: []letter.Option{
				letter.Attach("attach1", []byte{1, 2, 3}),
				letter.Attach("attach2", []byte{2, 3, 4, 5}, letter.ContentType("application/json")),
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
				letter.AttachFile("attach1", "./testdata/attachment.txt", letter.ContentType("text/html")),
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
			let, err := letter.Write(test.opts...)
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
			name: "To() recipieints",
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
			name: "CC() recipieints",
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
			name: "BCC() recipieints",
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
			name: "mixed recipieints",
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
			let, err := letter.Write(test.opts...)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, let.Recipients())
		})
	}
}

func TestLetter_RFC(t *testing.T) {
	let, err := letter.Write(
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
		letter.Attach("attach1", []byte("Attachment 1"), letter.ContentType("text/plain")),
		letter.Attach("attach2", []byte("<p>Attachment 2</p>"), letter.ContentType("text/html")),
	)
	assert.Nil(t, err)

	expected := join(
		"MIME-Version: 1.0",
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

	assert.Equal(t, expected, let.RFC())
	assert.Equal(t, expected, let.String())
}

func join(lines ...string) string {
	return strings.Join(lines, "\r\n")
}

func boundary(i int) string {
	return fmt.Sprintf("%064d", i+1)
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
