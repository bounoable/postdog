package cursor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bounoable/postdog/plugin/archive"
)

var (
	// ErrClosed means the cursor has been closed already.
	ErrClosed = errors.New("cursor closed")
)

// Cursor is a archive.Mail cursor.
type Cursor struct {
	mux    sync.RWMutex
	mails  []archive.Mail
	index  int
	closed bool
}

// New returns a new *Cursor, initialized with the provided mails.
func New(mails ...archive.Mail) *Cursor {
	return &Cursor{mails: mails, index: -1}
}

// All returns the remaining archive.Mails from the Cursor and calls cur.Close(ctx) afterwards.
//
// It fails with ErrClosed if cur.Close() has been called before.
func (cur *Cursor) All(ctx context.Context) ([]archive.Mail, error) {
	if cur.isClosed() {
		return nil, ErrClosed
	}

	var mails []archive.Mail
	for cur.Next(ctx) {
		mails = append(mails, cur.Current())
	}

	if err := cur.Close(ctx); err != nil {
		return mails, fmt.Errorf("close: %w", err)
	}

	return mails, nil
}

func (cur *Cursor) isClosed() bool {
	cur.mux.RLock()
	defer cur.mux.RUnlock()
	return cur.closed
}

// Next tries to advance the Cursor to the next mail. It returns false if the
// Cursor is already at the end, otherwise it returns true.
func (cur *Cursor) Next(ctx context.Context) bool {
	cur.mux.Lock()
	defer cur.mux.Unlock()
	if cur.index+1 >= len(cur.mails) {
		return false
	}
	cur.index++
	return true
}

// Err always returns nil.
func (cur *Cursor) Err() error {
	return nil
}

// Current returns the current archive.Mail.
func (cur *Cursor) Current() archive.Mail {
	cur.mux.RLock()
	defer cur.mux.RUnlock()
	if cur.index < 0 {
		return archive.Mail{}
	}
	return cur.mails[cur.index]
}

// Push adds mails to the end of the Cursor. It fails with ErrClosed
// if cur.Close() has been called before.
func (cur *Cursor) Push(mails ...archive.Mail) error {
	if cur.isClosed() {
		return ErrClosed
	}
	cur.mux.Lock()
	defer cur.mux.Unlock()
	cur.mails = append(cur.mails, mails...)
	return nil
}

// Close closes the Cursor and puts it in an unusable state. It fails with
// ErrClosed if cur.Close() has been called before.
func (cur *Cursor) Close(ctx context.Context) error {
	if cur.isClosed() {
		return ErrClosed
	}
	cur.mux.Lock()
	defer cur.mux.Unlock()
	cur.closed = true
	return nil
}
