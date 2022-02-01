package memory

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"sort"
	"strings"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/archive"
	"github.com/bounoable/postdog/plugin/archive/cursor"
	"github.com/bounoable/postdog/plugin/archive/query"
	"github.com/google/uuid"
)

// Store is an in-memory mail store.
type Store struct {
	mails []archive.Mail
}

// NewStore returns a new in-memory store.
func NewStore() *Store {
	return &Store{}
}

// Insert inserts m into s.
func (s *Store) Insert(ctx context.Context, m archive.Mail) error {
	s.mails = append(s.mails, m)
	return nil
}

// Find returns the archive.Mail with the given id.
func (s *Store) Find(ctx context.Context, id uuid.UUID) (archive.Mail, error) {
	for _, m := range s.mails {
		if m.ID() == id {
			return m, nil
		}
	}
	return archive.Mail{}, fmt.Errorf("find mail %s: %w", id, archive.ErrNotFound)
}

// Query returns a query.Cursor that returns the mails in the Store that match the query.Query q.
func (s *Store) Query(_ context.Context, q query.Query) (archive.Cursor, error) {
	var mails []archive.Mail
	for _, m := range s.mails {
		if filter(m, q) {
			mails = append(mails, m)
		}
	}
	mails = sortMails(mails, q)
	mails = paginate(mails, q)
	return cursor.New(mails...), nil
}

// Remove removes the given archive.Mail m from the Store s.
func (s *Store) Remove(_ context.Context, m archive.Mail) error {
	for i, c := range s.mails {
		if c.ID() == m.ID() {
			s.mails = append(s.mails[:i], s.mails[i+1:]...)
			return nil
		}
	}
	return nil
}

func filter(pm archive.Mail, q query.Query) bool {
	m := archive.ExpandMail(pm)

	if len(q.From) > 0 {
		var found bool
		for _, addr := range q.From {
			if addr == m.From() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(q.Recipients) > 0 {
		if !containsAnyAddress(m.Recipients(), q.Recipients) {
			return false
		}
	}

	if len(q.To) > 0 {
		if !containsAnyAddress(m.To(), q.To) {
			return false
		}
	}

	if len(q.CC) > 0 {
		if !containsAnyAddress(m.CC(), q.CC) {
			return false
		}
	}

	if len(q.BCC) > 0 {
		if !containsAnyAddress(m.BCC(), q.BCC) {
			return false
		}
	}

	// if len(q.RFC) > 0 {
	// 	if !containsAnySubstring(m.RFC(), q.RFC) {
	// 		return false
	// 	}
	// }

	// if len(q.Texts) > 0 {
	// 	if !containsAnySubstring(m.Text(), q.Texts) {
	// 		return false
	// 	}
	// }

	// if len(q.HTML) > 0 {
	// 	if !containsAnySubstring(m.HTML(), q.HTML) {
	// 		return false
	// 	}
	// }

	if len(q.Subjects) > 0 {
		if !containsAnySubstring(m.Subject(), q.Subjects) {
			return false
		}
	}

	// if len(q.SendErrors) > 0 {
	// 	if !containsAnySubstring(m.SendError(), q.SendErrors) {
	// 		return false
	// 	}
	// }

	attachments := m.Attachments()
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

			if want.Name == "" && want.Address == given.Address {
				return true
			}

			if want.Address == "" && want.Name == given.Name {
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

func sortMails(mails []archive.Mail, q query.Query) []archive.Mail {
	if q.Sorting == query.SortAny {
		return mails
	}

	sort.Slice(mails, func(a, b int) bool {
		mailA := archive.ExpandMail(mails[a])
		mailB := archive.ExpandMail(mails[b])

		switch q.Sorting {
		case query.SortSendTime:
			return compareTime(mailA.SentAt(), mailB.SentAt(), q.SortDirection == query.SortDesc)
		case query.SortSubject:
			r := mailA.Subject() < mailB.Subject()
			if q.SortDirection == query.SortDesc {
				return !r
			}
			return r
		default:
			return true
		}
	})

	return mails
}

func compareTime(a, b time.Time, desc bool) bool {
	if desc {
		return a.After(b)
	}
	return a.Before(b)
}

func paginate(mails []archive.Mail, q query.Query) []archive.Mail {
	if len(mails) == 0 {
		return mails
	}
	if q.Pagination.Page == 0 {
		return mails
	}
	if q.Pagination.PerPage == 0 {
		return nil
	}
	start := (q.Pagination.Page - 1) * q.Pagination.PerPage
	end := q.Pagination.Page * q.Pagination.PerPage
	if end > len(mails) {
		end = len(mails)
	}
	return mails[start:end]
}
