package letter

import (
	"net/mail"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachment_Map(t *testing.T) {
	tests := []struct {
		name string
		give Attachment
		opts []MapOption
		want func(Attachment) map[string]interface{}
	}{
		{
			name: "default",
			give: Attachment{
				filename:    "at1",
				content:     []byte{1, 2, 3},
				contentType: "text/plain",
			},
			want: func(at Attachment) map[string]interface{} {
				return map[string]interface{}{
					"filename":    "at1",
					"content":     []byte{1, 2, 3},
					"size":        3,
					"contentType": "text/plain",
					"header":      (map[string][]string)(at.header),
				}
			},
		},
		{
			name: "without content",
			give: Attachment{
				filename:    "at1",
				content:     []byte{1, 2, 3},
				contentType: "text/plain",
			},
			want: func(at Attachment) map[string]interface{} {
				return map[string]interface{}{
					"filename":    "at1",
					"content":     []byte{},
					"size":        3,
					"contentType": "text/plain",
					"header":      (map[string][]string)(at.header),
				}
			},
			opts: []MapOption{
				WithoutAttachmentContent(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.give.Map(tt.opts...)
			assert.Equal(t, tt.want(tt.give), m)
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
				Text("Hello."),
				HTML("<p>Hello.</p>"),
				Attach("at1", []byte{1, 2, 3}, AttachmentType("text/plain")),
				RFC("rfc body"),
			},
			want: func(l Letter) map[string]interface{} {
				return map[string]interface{}{
					"from": map[string]interface{}{
						"name":    "Bob Belcher",
						"address": "bob@example.com",
					},
					"recipients": []interface{}{
						map[string]interface{}{
							"name":    "Linda Belcher",
							"address": "linda@example.com",
						},
						map[string]interface{}{
							"name":    "Tina Belcher",
							"address": "tina@example.com",
						},
						map[string]interface{}{
							"name":    "Gene Belcher",
							"address": "gene@example.com",
						},
					},
					"to": []interface{}{
						map[string]interface{}{
							"name":    "Linda Belcher",
							"address": "linda@example.com",
						},
					},
					"cc": []interface{}{
						map[string]interface{}{
							"name":    "Tina Belcher",
							"address": "tina@example.com",
						},
					},
					"bcc": []interface{}{
						map[string]interface{}{
							"name":    "Gene Belcher",
							"address": "gene@example.com",
						},
					},
					"replyTo": []interface{}{
						map[string]interface{}{
							"name":    "Louise Belcher",
							"address": "louise@example.com",
						},
					},
					"subject": "Hi.",
					"text":    "Hello.",
					"html":    "<p>Hello.</p>",
					"attachments": []interface{}{
						map[string]interface{}{
							"filename":    "at1",
							"content":     []byte{1, 2, 3},
							"size":        3,
							"contentType": "text/plain",
							"header":      (map[string][]string)(l.attachments[0].header),
						},
					},
					"rfc": "rfc body",
				}
			},
		},
		{
			name: "without attachment content",
			letterOpts: []Option{
				From("Bob Belcher", "bob@example.com"),
				To("Linda Belcher", "linda@example.com"),
				CC("Tina Belcher", "tina@example.com"),
				BCC("Gene Belcher", "gene@example.com"),
				ReplyTo("Louise Belcher", "louise@example.com"),
				Subject("Hi."),
				Text("Hello."),
				HTML("<p>Hello.</p>"),
				Attach("at1", []byte{1, 2, 3}, AttachmentType("text/plain")),
				RFC("rfc body"),
			},
			opts: []MapOption{
				WithoutAttachmentContent(),
			},
			want: func(l Letter) map[string]interface{} {
				return map[string]interface{}{
					"from": map[string]interface{}{
						"name":    "Bob Belcher",
						"address": "bob@example.com",
					},
					"recipients": []interface{}{
						map[string]interface{}{
							"name":    "Linda Belcher",
							"address": "linda@example.com",
						},
						map[string]interface{}{
							"name":    "Tina Belcher",
							"address": "tina@example.com",
						},
						map[string]interface{}{
							"name":    "Gene Belcher",
							"address": "gene@example.com",
						},
					},
					"to": []interface{}{
						map[string]interface{}{
							"name":    "Linda Belcher",
							"address": "linda@example.com",
						},
					},
					"cc": []interface{}{
						map[string]interface{}{
							"name":    "Tina Belcher",
							"address": "tina@example.com",
						},
					},
					"bcc": []interface{}{
						map[string]interface{}{
							"name":    "Gene Belcher",
							"address": "gene@example.com",
						},
					},
					"replyTo": []interface{}{
						map[string]interface{}{
							"name":    "Louise Belcher",
							"address": "louise@example.com",
						},
					},
					"subject": "Hi.",
					"text":    "Hello.",
					"html":    "<p>Hello.</p>",
					"attachments": []interface{}{
						map[string]interface{}{
							"filename":    "at1",
							"content":     []byte{},
							"size":        3,
							"contentType": "text/plain",
							"header":      (map[string][]string)(l.attachments[0].header),
						},
					},
					"rfc": "rfc body",
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

func TestLetter_Parse(t *testing.T) {
	tests := []struct {
		name   string
		give   map[string]interface{}
		assert func(*testing.T, Letter)
	}{
		{
			name: "default",
			give: map[string]interface{}{
				"from": map[string]interface{}{
					"name":    "Bob Belcher",
					"address": "bob@example.com",
				},
				"recipients": []interface{}{
					map[string]interface{}{
						"name":    "Linda Belcher",
						"address": "linda@example.com",
					},
					map[string]interface{}{
						"name":    "Tina Belcher",
						"address": "tina@example.com",
					},
					map[string]interface{}{
						"name":    "Gene Belcher",
						"address": "gene@example.com",
					},
				},
				"to": []interface{}{
					map[string]interface{}{
						"name":    "Linda Belcher",
						"address": "linda@example.com",
					},
				},
				"cc": []interface{}{
					map[string]interface{}{
						"name":    "Tina Belcher",
						"address": "tina@example.com",
					},
				},
				"bcc": []interface{}{
					map[string]interface{}{
						"name":    "Gene Belcher",
						"address": "gene@example.com",
					},
				},
				"replyTo": []interface{}{
					map[string]interface{}{
						"name":    "Louise Belcher",
						"address": "louise@example.com",
					},
				},
				"subject": "Hi.",
				"text":    "Hello.",
				"html":    "<p>Hello.</p>",
				"attachments": []interface{}{
					map[string]interface{}{
						"filename":    "at1",
						"content":     []byte{1, 2, 3},
						"size":        3,
						"contentType": "text/plain",
						"header": map[string][]string{
							"key1": {"val"},
							"key2": {"val1", "val2"},
						},
					},
				},
				"rfc": "rfc body",
			},
			assert: func(t *testing.T, l Letter) {
				assert.Equal(t, mail.Address{Name: "Bob Belcher", Address: "bob@example.com"}, l.From())
				assert.Equal(t, []mail.Address{
					{Name: "Linda Belcher", Address: "linda@example.com"},
					{Name: "Tina Belcher", Address: "tina@example.com"},
					{Name: "Gene Belcher", Address: "gene@example.com"},
				}, l.Recipients())
				assert.Equal(t, []mail.Address{
					{Name: "Linda Belcher", Address: "linda@example.com"},
				}, l.To())
				assert.Equal(t, []mail.Address{
					{Name: "Tina Belcher", Address: "tina@example.com"},
				}, l.CC())
				assert.Equal(t, []mail.Address{
					{Name: "Gene Belcher", Address: "gene@example.com"},
				}, l.BCC())
				assert.Equal(t, []mail.Address{
					{Name: "Louise Belcher", Address: "louise@example.com"},
				}, l.ReplyTo())
				assert.Equal(t, "Hi.", l.Subject())
				assert.Equal(t, "Hello.", l.Text())
				assert.Equal(t, "<p>Hello.</p>", l.HTML())
				assert.Equal(t, "rfc body", l.RFC())
				assert.Equal(t, []Attachment{
					{
						filename:    "at1",
						content:     []byte{1, 2, 3},
						contentType: "text/plain",
						size:        0, // because non-0 means override actual size
						header: textproto.MIMEHeader{
							"key1": {"val"},
							"key2": {"val1", "val2"},
						},
					},
				}, l.Attachments())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l Letter
			l.Parse(tt.give)
			tt.assert(t, l)
		})
	}
}
