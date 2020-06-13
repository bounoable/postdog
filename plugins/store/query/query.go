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
	SentAt SentAtFilter
}

// SentAtFilter ...
type SentAtFilter struct {
	Before time.Time
	After  time.Time
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
