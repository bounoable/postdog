package letter_test

import (
	"bytes"
	"errors"
	"net/mail"
	"testing"

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
				assert.Equal(t, `application/octet-stream; name="attach1"`, at.Header().Get("Content-Type"))
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
				assert.Equal(t, `text/plain; charset=utf-8; name="attach1.txt"`, at.Header().Get("Content-Type"))
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
				assert.Equal(t, `application/json; name="attach1"`, at.Header().Get("Content-Type"))
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
				assert.Equal(t, `application/octet-stream; name="attach1"`, at.Header().Get("Content-Type"))

				at = l.Attachments()[1]
				assert.Equal(t, "attach2", at.Filename())
				assert.Equal(t, 4, at.Size())
				assert.Equal(t, []byte{2, 3, 4, 5}, at.Content())
				assert.Equal(t, "application/json", at.ContentType())
				assert.Equal(t, `application/json; name="attach2"`, at.Header().Get("Content-Type"))
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
				assert.Equal(t, `application/octet-stream; name="attach1"`, at.Header().Get("Content-Type"))
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
				assert.Equal(t, `text/plain; charset=utf-8; name="attach1"`, at.Header().Get("Content-Type"))
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
				assert.Equal(t, `text/html; name="attach1"`, at.Header().Get("Content-Type"))
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

type mockReader struct {
	readFn func([]byte) (int, error)
}

func (mr mockReader) Read(b []byte) (int, error) {
	return mr.readFn(b)
}

var mockError = errors.New("mock error")
