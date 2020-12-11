package letter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachment_Map(t *testing.T) {
	tests := []struct {
		name string
		give Attachment
		opts []MapOption
		want map[string]interface{}
	}{
		{
			name: "default",
			give: Attachment{
				filename:    "at1",
				content:     []byte{1, 2, 3},
				contentType: "text/plain",
			},
			want: map[string]interface{}{
				"filename":    "at1",
				"content":     []byte{1, 2, 3},
				"size":        3,
				"contentType": "text/plain",
			},
		},
		{
			name: "without content",
			give: Attachment{
				filename:    "at1",
				content:     []byte{1, 2, 3},
				contentType: "text/plain",
			},
			want: map[string]interface{}{
				"filename":    "at1",
				"content":     []byte{},
				"size":        3,
				"contentType": "text/plain",
			},
			opts: []MapOption{
				WithoutAttachmentContent(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.give.Map(tt.opts...)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestAttachment_Parse(t *testing.T) {
	tests := []struct {
		name   string
		give   map[string]interface{}
		assert func(*testing.T, Attachment)
	}{
		{
			name: "default",
			give: map[string]interface{}{
				"filename":    "at1",
				"content":     []byte{1, 2, 3},
				"size":        3,
				"contentType": "text/plain",
			},
			assert: func(t *testing.T, a Attachment) {
				assert.Equal(t, "at1", a.Filename())
				assert.Equal(t, []byte{1, 2, 3}, a.Content())
				assert.Equal(t, "text/plain", a.ContentType())
				assert.Equal(t, 3, a.Size())
			},
		},
		{
			name: "without content",
			give: map[string]interface{}{
				"filename":    "at1",
				"size":        3,
				"contentType": "text/plain",
			},
			assert: func(t *testing.T, a Attachment) {
				assert.Equal(t, "at1", a.Filename())
				assert.Equal(t, []byte(nil), a.Content())
				assert.Equal(t, "text/plain", a.ContentType())
				assert.Equal(t, 3, a.Size())
			},
		},
		{
			name: "without size",
			give: map[string]interface{}{
				"filename":    "at1",
				"content":     []byte{1, 2, 3},
				"contentType": "text/plain",
			},
			assert: func(t *testing.T, a Attachment) {
				assert.Equal(t, "at1", a.Filename())
				assert.Equal(t, []byte{1, 2, 3}, a.Content())
				assert.Equal(t, "text/plain", a.ContentType())
				assert.Equal(t, 3, a.Size())
			},
		},
		{
			name: "content <-> size mismatch",
			give: map[string]interface{}{
				"filename":    "at1",
				"content":     []byte{1, 2, 3},
				"size":        5,
				"contentType": "text/plain",
			},
			assert: func(t *testing.T, a Attachment) {
				assert.Equal(t, "at1", a.Filename())
				assert.Equal(t, []byte{1, 2, 3}, a.Content())
				assert.Equal(t, "text/plain", a.ContentType())
				assert.Equal(t, 5, a.Size())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var at Attachment
			at.Parse(tt.give)
			tt.assert(t, at)
		})
	}
}

func TestLetter_Map(t *testing.T) {
	tests := []struct {
		name       string
		letterOpts []Option
		opts       []MapOption
		want       func(Letter) map[string]interface{}
	}{
		{
			name: "default",
			letterOpts: []Option{
				From("Bob Belcher", "bob@example.com"),
				To("Linda Belcher", "linda@example.com"),
				CC("Tina Belcher", "tina@example.com"),
				BCC("Gene Belcher", "gene@example.com"),
				ReplyTo("Louise Belcher", "louise@example.com"),
				Subject("Hi."),
				Text("Hello"),
				HTML("<p>Hello.</p>"),
				Attach("attach1", []byte{1, 2, 3}, ContentType("text/plain")),
			},
			want: func(l Letter) map[string]interface{} {
				return map[string]interface{}{
					"from":        l.From(),
					"recipients":  l.Recipients(),
					"to":          l.To(),
					"cc":          l.CC(),
					"bcc":         l.BCC(),
					"replyTo":     l.ReplyTo(),
					"subject":     l.Subject(),
					"text":        l.Text(),
					"html":        l.HTML(),
					"attachments": mapAttachments(l.Attachments()),
				}
			},
		},
		{
			name: "without attachment contents",
			letterOpts: []Option{
				From("Bob Belcher", "bob@example.com"),
				To("Linda Belcher", "linda@example.com"),
				CC("Tina Belcher", "tina@example.com"),
				BCC("Gene Belcher", "gene@example.com"),
				ReplyTo("Louise Belcher", "louise@example.com"),
				Subject("Hi."),
				Text("Hello"),
				HTML("<p>Hello.</p>"),
				Attach("attach1", []byte{1, 2, 3}, ContentType("text/plain")),
			},
			opts: []MapOption{
				WithoutAttachmentContent(),
			},
			want: func(l Letter) map[string]interface{} {
				return map[string]interface{}{
					"from":        l.From(),
					"recipients":  l.Recipients(),
					"to":          l.To(),
					"cc":          l.CC(),
					"bcc":         l.BCC(),
					"replyTo":     l.ReplyTo(),
					"subject":     l.Subject(),
					"text":        l.Text(),
					"html":        l.HTML(),
					"attachments": mapAttachments(l.Attachments(), WithoutAttachmentContent()),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Write(tt.letterOpts...)
			assert.Equal(t, tt.want(l), l.Map(tt.opts...))
		})
	}
}

func mapAttachments(ats []Attachment, opts ...MapOption) []map[string]interface{} {
	mapped := make([]map[string]interface{}, len(ats))
	for i, at := range ats {
		mapped[i] = at.Map(opts...)
	}
	return mapped
}
