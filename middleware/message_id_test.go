package middleware_test

import (
	"context"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/letter/rfc"
	"github.com/bounoable/postdog/middleware"
	"github.com/stretchr/testify/assert"
)

func TestMessageID(t *testing.T) {
	var factory rfc.MessageIDFactory = rfc.MessageIDFunc(func(rfc.Mail) string {
		return "<foo@example.com>"
	})

	mw := middleware.MessageID(factory)
	_, l, err := postdog.ApplyMiddleware(context.Background(), letter.Write(), mw)

	assert.Nil(t, err)

	body := l.RFC()
	assert.Contains(t, body, "Message-ID: <foo@example.com>")
}
