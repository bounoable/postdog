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

// Query ...
type Query struct {
	SentAt SentAtFilter
}

// SentAtFilter ...
type SentAtFilter struct {
	Before time.Time
	After  time.Time
}

// Cursor ...
type Cursor interface {
	Next(context.Context) bool
	Current() store.Letter
	Close(context.Context) error
	Err() error
}
