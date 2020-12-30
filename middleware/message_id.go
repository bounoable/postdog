package middleware

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/letter/rfc"
)

// MessageID returns a Middleware that specifies the rfc.MessageIDFactory
// that is used for generating unique Message-IDs for the RFC body of the Mail.
func MessageID(factory rfc.MessageIDFactory) postdog.MiddlewareFunc {
	opt := rfc.WithMessageIDFactory(factory)
	return func(ctx context.Context, m postdog.Mail, next postdog.NextMiddleware) (postdog.Mail, error) {
		l := letter.Expand(m)
		cfg := l.RFCConfig()
		opt(&cfg)
		l = l.WithRFCConfig(cfg)
		return next(ctx, l)
	}
}
