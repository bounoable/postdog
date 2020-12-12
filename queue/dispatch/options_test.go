package dispatch_test

import (
	"testing"
	"time"

	"github.com/bounoable/postdog/queue/dispatch"
	"github.com/bounoable/postdog/send"
	"github.com/stretchr/testify/assert"
)

func TestSendOptions(t *testing.T) {
	cfg := dispatch.Configure(dispatch.SendOptions(send.Use("a"), send.Use("b")))
	assert.Len(t, cfg.SendOptions, 2)
	dispatch.SendOptions(send.Use("a"), send.Use("b"))(&cfg)
	assert.Len(t, cfg.SendOptions, 4)
}

func TestTimeout(t *testing.T) {
	cfg := dispatch.Configure(dispatch.Timeout(time.Millisecond * 2371))
	assert.Equal(t, time.Millisecond*2371, cfg.Timeout)
}
