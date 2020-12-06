package dispatch_test

import (
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/queue/dispatch"
	"github.com/stretchr/testify/assert"
)

func TestSendOptions(t *testing.T) {
	cfg := dispatch.Configure(dispatch.SendOptions(postdog.Use("a"), postdog.Use("b")))
	assert.Len(t, cfg.SendOptions, 2)
	dispatch.SendOptions(postdog.Use("a"), postdog.Use("b"))(&cfg)
	assert.Len(t, cfg.SendOptions, 4)
}
