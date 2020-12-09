package archive

import (
	"context"
	"fmt"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/plugin/archive/query"
)

//go:generate mockgen -source=archive.go -destination=./mocks/archive.go

// Store is the underlying store for the Mails.
type Store interface {
	Insert(context.Context, postdog.Mail) error
	Query(context.Context, query.Query) (query.Cursor, error)
}

// Printer is the logger interface.
type Printer interface {
	Print(...interface{})
}

// Option is an archive option.
type Option func(*config)

type config struct {
	logger Printer
}

// New creates the archive plugin.
func New(s Store, opts ...Option) postdog.Plugin {
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	return postdog.Plugin{
		postdog.WithHook(postdog.AfterSend, postdog.ListenerFunc(func(
			ctx context.Context,
			_ postdog.Hook,
			pm postdog.Mail,
		) {
			sendError := postdog.SendError(ctx)
			sentAt := postdog.SendTime(ctx)

			var errMsg string
			if sendError != nil {
				errMsg = sendError.Error()
			}

			m := ExpandMail(pm).
				WithSendError(errMsg).
				WithSendTime(sentAt)

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

func (cfg *config) logInsertError(err error) {
	if cfg.logger != nil {
		cfg.logger.Print(fmt.Sprintf("Failed to insert mail into store: %s\n", err.Error()))
	}
}
