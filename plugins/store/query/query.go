package query

//go:generate mockgen -source=query.go -destination=./mock_query/query.go

import (
	"context"
	"time"

	"github.com/bounoable/postdog/plugins/store"
)

// Repository ...
type Repository interface {
	Query(context.Context, Query) (Cursor, error)
}

// Cursor ...
type Cursor interface {
	Next(context.Context) bool
	Current() store.Letter
	Close(context.Context) error
	Err() error
}

// Query ...
type Query struct {
	SentAt     SentAtFilter
	Subjects   []string
	From       []string
	To         []string
	CC         []string
	BCC        []string
	Attachment AttachmentFilter
}

// SentAtFilter ...
type SentAtFilter struct {
	Before time.Time
	After  time.Time
}

// AttachmentFilter ...
type AttachmentFilter struct {
	Names        []string
	ContentTypes []string
	Size         AttachmentSizeFilter
}

// AttachmentSizeFilter ...
type AttachmentSizeFilter struct {
	Exact  []int
	Ranges [][2]int
}

// New ...
func New(opts ...Option) Query {
	var q Query
	for _, opt := range opts {
		opt(&q)
	}
	return q
}

// Run ...
func Run(ctx context.Context, repo Repository, opts ...Option) (Cursor, error) {
	return repo.Query(ctx, New(opts...))
}

// Option ...
type Option func(*Query)

// SentBefore ...
func SentBefore(t time.Time) Option {
	return func(q *Query) {
		q.SentAt.Before = t
	}
}

// SentAfter ...
func SentAfter(t time.Time) Option {
	return func(q *Query) {
		q.SentAt.After = t
	}
}

// SentInBetween ...
func SentInBetween(l, r time.Time) Option {
	return func(q *Query) {
		SentBefore(r)(q)
		SentAfter(l)(q)
	}
}

// SentBetween ...
func SentBetween(l, r time.Time) Option {
	return SentInBetween(l.Add(-time.Nanosecond), r.Add(time.Nanosecond))
}

// Subject ...
func Subject(s ...string) Option {
	return func(q *Query) {
		q.Subjects = append(q.Subjects, s...)
	}
}

// From ...
func From(f ...string) Option {
	return func(q *Query) {
		q.From = append(q.From, f...)
	}
}

// To ...
func To(to ...string) Option {
	return func(q *Query) {
		q.To = append(q.To, to...)
	}
}

// CC ...
func CC(cc ...string) Option {
	return func(q *Query) {
		q.CC = append(q.CC, cc...)
	}
}

// BCC ...
func BCC(bcc ...string) Option {
	return func(q *Query) {
		q.BCC = append(q.BCC, bcc...)
	}
}

// AttachmentName ...
func AttachmentName(n ...string) Option {
	return func(q *Query) {
		q.Attachment.Names = append(q.Attachment.Names, n...)
	}
}

// AttachmentContentType ...
func AttachmentContentType(ct ...string) Option {
	return func(q *Query) {
		q.Attachment.ContentTypes = append(q.Attachment.ContentTypes, ct...)
	}
}

// AttachmentSize ...
func AttachmentSize(s ...int) Option {
	return func(q *Query) {
		q.Attachment.Size.Exact = append(q.Attachment.Size.Exact, s...)
	}
}

// AttachmentSizeRange ...
func AttachmentSizeRange(min, max int) Option {
	if min > max {
		min, max = max, min
	}

	return func(q *Query) {
		q.Attachment.Size.Ranges = append(q.Attachment.Size.Ranges, [2]int{min, max})
	}
}
