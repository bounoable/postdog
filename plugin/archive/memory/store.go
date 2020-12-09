package memory

import (
	"bytes"
	"context"
	"net/mail"
	"strings"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/archive/cursor"
	"github.com/bounoable/postdog/plugin/archive/query"
)

type Store struct {
	mails []postdog.Mail
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Insert(ctx context.Context, m postdog.Mail) error {
	s.mails = append(s.mails, m)
	return nil
}

func (s *Store) Query(_ context.Context, q query.Query) (query.Cursor, error) {
	var mails []postdog.Mail
	for _, m := range s.mails {
		if filter(m, q) {
			mails = append(mails, m)
		}
	}
	return cursor.New(mails...), nil
}

func filter(pm postdog.Mail, q query.Query) bool {
	l := letter.Expand(pm)

	if len(q.From) > 0 {
		var found bool
		for _, addr := range q.From {
			if addr == l.From() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(q.Recipients) > 0 {
		if !containsAnyAddress(l.Recipients(), q.Recipients) {
			return false
		}
	}

	if len(q.To) > 0 {
		if !containsAnyAddress(l.To(), q.To) {
			return false
		}
	}

	if len(q.CC) > 0 {
		if !containsAnyAddress(l.CC(), q.CC) {
			return false
		}
	}

	if len(q.BCC) > 0 {
		if !containsAnyAddress(l.BCC(), q.BCC) {
			return false
		}
	}

	if len(q.RFC) > 0 {
		if !containsAnySubstring(l.RFC(), q.RFC) {
			return false
		}
	}

	if len(q.Texts) > 0 {
		if !containsAnySubstring(l.Text(), q.Texts) {
			return false
		}
	}

	if len(q.HTML) > 0 {
		if !containsAnySubstring(l.HTML(), q.HTML) {
			return false
		}
	}

	if len(q.Subjects) > 0 {
		if !containsAnySubstring(l.Subject(), q.Subjects) {
			return false
		}
	}

	attachments := l.Attachments()
	if len(q.Attachment.Filenames) > 0 {
		if !containsAnyAttachmentFilename(attachments, q.Attachment.Filenames) {
			return false
		}
	}

	if len(q.Attachment.Size.Exact) > 0 {
		if !containsAnyAttachmentSize(attachments, q.Attachment.Size.Exact) {
			return false
		}
	}

	if len(q.Attachment.Size.Ranges) > 0 {
		if !containsAnyAttachmentSizeRange(attachments, q.Attachment.Size.Ranges) {
			return false
		}
	}

	if len(q.Attachment.ContentTypes) > 0 {
		if !containsAnyAttachmentContentType(attachments, q.Attachment.ContentTypes) {
			return false
		}
	}

	if len(q.Attachment.Contents) > 0 {
		if !containsAnyAttachmentContent(attachments, q.Attachment.Contents) {
			return false
		}
	}

	return true
}

func containsAnyAddress(addrs []mail.Address, search []mail.Address) bool {
	for _, given := range addrs {
		for _, want := range search {
			if given == want {
				return true
			}
		}
	}
	return false
}

func containsAnySubstring(s string, ss []string) bool {
	for _, sub := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func containsAnyAttachmentFilename(ats []letter.Attachment, filenames []string) bool {
	for _, at := range ats {
		for _, name := range filenames {
			if name == at.Filename() {
				return true
			}
		}
	}
	return false
}

func containsAnyAttachmentSize(ats []letter.Attachment, sizes []int) bool {
	for _, at := range ats {
		for _, size := range sizes {
			if at.Size() == size {
				return true
			}
		}
	}
	return false
}

func containsAnyAttachmentSizeRange(ats []letter.Attachment, ranges [][2]int) bool {
	for _, at := range ats {
		for _, r := range ranges {
			size := at.Size()
			if r[0] <= size && r[1] >= size {
				return true
			}
		}
	}
	return false
}

func containsAnyAttachmentContentType(ats []letter.Attachment, cts []string) bool {
	for _, at := range ats {
		for _, ct := range cts {
			if at.ContentType() == ct {
				return true
			}
		}
	}
	return false
}

func containsAnyAttachmentContent(ats []letter.Attachment, contents [][]byte) bool {
	for _, at := range ats {
		atc := at.Content()
		for _, content := range contents {
			if bytes.Equal(content, atc) {
				return true
			}
		}
	}
	return false
}
