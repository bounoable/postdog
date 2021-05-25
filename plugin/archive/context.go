package archive

import (
	"context"

	"github.com/google/uuid"
)

const (
	ctxMailID = ctxKey("mail_id")
)

type ctxKey string

// WithMailID returns a new Context that carries the given UUID. Mails that are
// archived with that Context will use that UUID as their UUID when stored in a
// database.
func WithMailID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxMailID, id)
}

// MailIDFromContext returns the mail UUID from the given Context, or uuid.Nil
// if the Context has no mail UUID.
func MailIDFromContext(ctx context.Context) uuid.UUID {
	val := ctx.Value(ctxMailID)
	if val == nil {
		return uuid.Nil
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return id
}
