// Package query provides functions to query letters from a store.
package query

//go:generate mockgen -source=query.go -destination=./mock_query/query.go

import (
	"context"
	"time"

	"github.com/bounoable/postdog/plugin/store"
	"github.com/google/uuid"
)

const (
	// SortBySendDate sorts letters by the "SentAt" field.
	SortBySendDate = Sorting(iota)
)

const (
	// SortAsc sorts letters in ascending order.
	SortAsc = SortDirection(iota)
	// SortDesc sorts letters in descending order.
	SortDesc
)

// Repository is the query repository.
type Repository interface {
	Query(context.Context, Query) (Cursor, error)
	Get(context.Context, uuid.UUID) (store.Letter, error)
}

// Cursor is used to iterate over a stream of letters.
type Cursor interface {
	Next(context.Context) bool
	Current() store.Letter
	Close(context.Context) error
	Err() error
}

// Query defines the filters and options for a query.
type Query struct {
	SentAt     SentAtFilter
	Subjects   []string
	From       []string
	To         []string
	CC         []string
	BCC        []string
	Attachment AttachmentFilter
	Sort       SortConfig
}

// SentAtFilter filters letters by their send date.
type SentAtFilter struct {
	Before time.Time
	After  time.Time
}

// AttachmentFilter filters letters by their attachments.
type AttachmentFilter struct {
	Names        []string
	ContentTypes []string
	Size         AttachmentSizeFilter
}

// AttachmentSizeFilter filters letters by their attachment sizes.
type AttachmentSizeFilter struct {
	Exact  []int
	Ranges [][2]int
}

// SortConfig defines the sorting of queried letters.
type SortConfig struct {
	SortBy Sorting
	Dir    SortDirection
}

// Sorting is the sorting of letters.
type Sorting int

// SortDirection is the direction of the sorting.
type SortDirection int

// New builds and returns a new Query, configured by opts.
func New(opts ...Option) Query {
	var q Query
	for _, opt := range opts {
		opt(&q)
	}
	return q
}

// Run builds and runs a query against repo.
// Use opts to configure the query.
func Run(ctx context.Context, repo Repository, opts ...Option) (Cursor, error) {
	return repo.Query(ctx, New(opts...))
}

// Option is a query option.
type Option func(*Query)

// SentBefore includes only letters which have been sent before t.
func SentBefore(t time.Time) Option {
	return func(q *Query) {
		q.SentAt.Before = t
	}
}

// SentAfter includes only letters which have been sent after t.
func SentAfter(t time.Time) Option {
	return func(q *Query) {
		q.SentAt.After = t
	}
}

// SentInBetween includes only letters which have been sent between (l, r).
func SentInBetween(l, r time.Time) Option {
	return func(q *Query) {
		SentBefore(r)(q)
		SentAfter(l)(q)
	}
}

// SentBetween includes only letters which have been sent between [l, r].
func SentBetween(l, r time.Time) Option {
	return SentInBetween(l.Add(-time.Nanosecond), r.Add(time.Nanosecond))
}

// Subject includes only letters whose subject contains s.
func Subject(s ...string) Option {
	return func(q *Query) {
		q.Subjects = append(q.Subjects, s...)
	}
}

// From includes only letters whose sender's name or address contains f.
func From(f ...string) Option {
	return func(q *Query) {
		q.From = append(q.From, f...)
	}
}

// To includes only letters whose "To" recipient's name or address contains to.
func To(to ...string) Option {
	return func(q *Query) {
		q.To = append(q.To, to...)
	}
}

// CC includes only letters whose "CC" recipients' names or addresses contains cc.
func CC(cc ...string) Option {
	return func(q *Query) {
		q.CC = append(q.CC, cc...)
	}
}

// BCC includes only letters whose "BCC" recipients' names or addresses contains bcc.
func BCC(bcc ...string) Option {
	return func(q *Query) {
		q.BCC = append(q.BCC, bcc...)
	}
}

// AttachmentName includes only letters that have an attachment whose name contains n.
func AttachmentName(n ...string) Option {
	return func(q *Query) {
		q.Attachment.Names = append(q.Attachment.Names, n...)
	}
}

// AttachmentContentType includes only letters that have
// an attachment whose "Content-Type" header contains ct.
func AttachmentContentType(ct ...string) Option {
	return func(q *Query) {
		q.Attachment.ContentTypes = append(q.Attachment.ContentTypes, ct...)
	}
}

// AttachmentSize includes only letters that have an attachment whose filesize is exactly s.
func AttachmentSize(s ...int) Option {
	return func(q *Query) {
		q.Attachment.Size.Exact = append(q.Attachment.Size.Exact, s...)
	}
}

// AttachmentSizeRange includes only letters that have an attachment whose filesize is in the range [min, max].
func AttachmentSizeRange(min, max int) Option {
	if min > max {
		min, max = max, min
	}

	return func(q *Query) {
		q.Attachment.Size.Ranges = append(q.Attachment.Size.Ranges, [2]int{min, max})
	}
}

// Sort configures the sorting of the letters.
func Sort(by Sorting, dir SortDirection) Option {
	return func(q *Query) {
		q.Sort = SortConfig{
			SortBy: by,
			Dir:    dir,
		}
	}
}

// Find tries to retrieve the letter with the given id from repo.
// If it can't find the letter, a `LetterNotFoundError` is returned.
func Find(ctx context.Context, repo Repository, id uuid.UUID) (store.Letter, error) {
	let, err := repo.Get(ctx, id)
	if err != nil {
		return store.Letter{}, LetterNotFoundError{
			ID:  id,
			Err: err,
		}
	}
	return let, nil
}
