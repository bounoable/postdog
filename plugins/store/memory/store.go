package memory

import (
	"context"
	"net/mail"
	"sort"
	"strings"
	"sync"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugins/store"
	"github.com/bounoable/postdog/plugins/store/query"
	"github.com/bounoable/timefn"
)

// Store ...
type Store struct {
	mux     sync.RWMutex
	letters []store.Letter
}

// NewStore ...
func NewStore(letters ...store.Letter) *Store {
	return &Store{
		letters: letters,
	}
}

// Insert ...
func (s *Store) Insert(ctx context.Context, let store.Letter) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.letters = append(s.letters, let)
	return nil
}

// Query ...
func (s *Store) Query(ctx context.Context, q query.Query) (query.Cursor, error) {
	var letters []store.Letter

	for _, let := range s.letters {
		if filterLetter(let, q) {
			letters = append(letters, let)
		}
	}

	letters = sortLetters(letters, q.Sort)

	return newCursor(letters), nil
}

func filterLetter(let store.Letter, q query.Query) bool {
	if !q.SentAt.Before.IsZero() && timefn.SameOrAfter(let.SentAt, q.SentAt.Before) {
		return false
	}

	if !q.SentAt.After.IsZero() && timefn.SameOrBefore(let.SentAt, q.SentAt.After) {
		return false
	}

	if len(q.Subjects) > 0 && !filterSubject(let, q.Subjects) {
		return false
	}

	if len(q.From) > 0 && !filterFrom(let, q.From) {
		return false
	}

	if len(q.To) > 0 && !filterAddresses(let.To, q.To) {
		return false
	}

	if len(q.CC) > 0 && !filterAddresses(let.CC, q.CC) {
		return false
	}

	if len(q.BCC) > 0 && !filterAddresses(let.BCC, q.BCC) {
		return false
	}

	if !filterAttachments(let.Attachments, q.Attachment) {
		return false
	}

	return true
}

func filterSubject(let store.Letter, subjects []string) bool {
	for _, subject := range subjects {
		if strings.Contains(let.Subject, subject) {
			return true
		}
	}
	return false
}

func filterFrom(let store.Letter, from []string) bool {
	for _, f := range from {
		if strings.Contains(let.From.Name, f) || strings.Contains(let.From.Address, f) {
			return true
		}
	}
	return false
}

func filterAddresses(addresses []mail.Address, search []string) bool {
	for _, saddr := range search {
		for _, addr := range addresses {
			if strings.Contains(addr.Name, saddr) || strings.Contains(addr.Address, saddr) {
				return true
			}
		}
	}
	return false
}

func filterAttachments(attachments []letter.Attachment, filter query.AttachmentFilter) bool {
	if len(filter.Names) > 0 && !filterAttachmentNames(attachments, filter.Names) {
		return false
	}

	if len(filter.ContentTypes) > 0 && !filterAttachmentContentTypes(attachments, filter.ContentTypes) {
		return false
	}

	if len(filter.Size.Exact) > 0 && !filterAttachmentSizeExact(attachments, filter.Size.Exact) {
		return false
	}

	if len(filter.Size.Ranges) > 0 && !filterAttachmentSizeRanges(attachments, filter.Size.Ranges) {
		return false
	}

	return true
}

func filterAttachmentNames(attachments []letter.Attachment, names []string) bool {
	for _, attach := range attachments {
		for _, name := range names {
			if strings.Contains(attach.Filename, name) {
				return true
			}
		}
	}
	return false
}

func filterAttachmentContentTypes(attachments []letter.Attachment, cts []string) bool {
	for _, attach := range attachments {
		for _, ct := range cts {
			if strings.Contains(attach.ContentType, ct) {
				return true
			}
		}
	}
	return false
}

func filterAttachmentSizeExact(attachments []letter.Attachment, sizes []int) bool {
	for _, attach := range attachments {
		for _, size := range sizes {
			if len(attach.Content) == size {
				return true
			}
		}
	}
	return false
}

func filterAttachmentSizeRanges(attachments []letter.Attachment, ranges [][2]int) bool {
	for _, attach := range attachments {
		size := len(attach.Content)
		for _, r := range ranges {
			if size >= r[0] && size <= r[1] {
				return true
			}
		}
	}
	return false
}

func sortLetters(letters []store.Letter, cfg query.SortConfig) []store.Letter {
	switch cfg.SortBy {
	case query.SortBySendDate:
		return sortLettersBySendDate(letters, cfg.Dir)
	default:
		return letters
	}
}

func sortLettersBySendDate(letters []store.Letter, dir query.SortDirection) []store.Letter {
	sort.Slice(letters, func(a, b int) bool {
		switch dir {
		case query.SortAsc:
			return letters[a].SentAt.Before(letters[b].SentAt)
		case query.SortDesc:
			return letters[a].SentAt.After(letters[b].SentAt)
		default:
			return true
		}
	})
	return letters
}

func newCursor(letters []store.Letter) query.Cursor {
	return &cursor{
		letters: letters,
		current: -1,
	}
}

type cursor struct {
	letters []store.Letter
	current int
}

func (cur *cursor) Next(_ context.Context) bool {
	if cur.current+1 >= len(cur.letters) {
		return false
	}
	cur.current++
	return true
}

func (cur *cursor) Current() store.Letter {
	if len(cur.letters) <= cur.current {
		return store.Letter{}
	}
	return cur.letters[cur.current]
}

func (cur *cursor) Close(_ context.Context) error {
	return nil
}

func (cur *cursor) Err() error {
	return nil
}
