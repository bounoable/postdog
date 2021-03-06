package archive

//go:generate mockgen -source=archive.go -destination=./mocks/archive.go

import (
	"context"
	stdctx "context"
	"errors"
	"fmt"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/plugin/archive/query"
	"github.com/google/uuid"
)

var (
	// ErrNotFound means a mail could not be found in the Store.
	ErrNotFound = errors.New("mail not found")
)

// Store is the underlying store for the Mails.
type Store interface {
	// Insert inserts a Mail into the Store.
	Insert(stdctx.Context, Mail) error

	// Find returns the Mail with the given ID.
	Find(stdctx.Context, uuid.UUID) (Mail, error)

	// Query queries the Store using the given query.Query.
	Query(stdctx.Context, query.Query) (Cursor, error)

	// Remove removes the given Mail from the Store.
	Remove(stdctx.Context, Mail) error
}

// Cursor is a cursor archived Mails.
type Cursor interface {
	// Next advances the cursor to the next Mail.
	// Implementations should return true if the next call to Current() would
	// return a valid Mail, or false if the Cursor reached the end or if Next()
	// failed because of an error. In the latter case, Err() should return that error.
	Next(stdctx.Context) bool

	// Current returns the current Mail.
	Current() Mail

	// All returns the remaining Mails from the Cursor and calls cur.Close(ctx) afterwards.
	All(stdctx.Context) ([]Mail, error)

	// Err returns the current error that occurred during a previous Next() call.
	Err() error

	// Close closes the Cursor. Users must call Close() after using the Cursor if they don't call All().
	Close(stdctx.Context) error
}

// Printer is the logger interface.
type Printer interface {
	Print(...interface{})
}

// Option is an archive option.
type Option func(*config)

type config struct {
	logger        Printer
	insertTimeout time.Duration
}

// New creates the archive plugin.
func New(s Store, opts ...Option) postdog.Plugin {
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	return postdog.Plugin{
		postdog.WithHook(postdog.AfterSend, postdog.ListenerFunc(func(
			ctx stdctx.Context,
			_ postdog.Hook,
			pm postdog.Mail,
		) {
			sendError := postdog.SendError(ctx)
			sentAt := postdog.SendTime(ctx)

			var errMsg string
			if sendError != nil {
				errMsg = sendError.Error()
			}

			id := MailIDFromContext(ctx)
			if id == uuid.Nil {
				id = uuid.New()
			}

			m := ExpandMail(pm).
				WithID(id).
				WithSendError(errMsg).
				WithSendTime(sentAt)

			var cancel context.CancelFunc
			if cfg.insertTimeout == 0 {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), cfg.insertTimeout)
			}
			defer cancel()

			if err := s.Insert(ctx, m); err != nil {
				cfg.logInsertError(err)
			}
		})),
	}
}

// WithLogger returns an Option that sets the error logger.
func WithLogger(l Printer) Option {
	return func(cfg *config) {
		cfg.logger = l
	}
}

// InsertTimeout returns an Option that sets the timeout for inserts.
func InsertTimeout(d time.Duration) Option {
	return func(cfg *config) {
		cfg.insertTimeout = d
	}
}

func (cfg *config) logInsertError(err error) {
	if cfg.logger != nil {
		cfg.logger.Print(fmt.Sprintf("Failed to insert mail into store: %s\n", err.Error()))
	}
}
