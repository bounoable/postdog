package memory

import (
	"context"
	"sync"

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

	return newCursor(letters), nil
}

func filterLetter(let store.Letter, q query.Query) bool {
	if !q.SentAt.Before.IsZero() && timefn.SameOrAfter(let.SentAt, q.SentAt.Before) {
		return false
	}

	if !q.SentAt.After.IsZero() && timefn.SameOrBefore(let.SentAt, q.SentAt.After) {
		return false
	}

	return true
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
