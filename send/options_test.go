package send_test

import (
	"testing"
	"time"

	"github.com/bounoable/postdog/send"
	"github.com/stretchr/testify/assert"
)

func TestUse(t *testing.T) {
	var cfg send.Config
	send.Use("test")(&cfg)
	assert.Equal(t, "test", cfg.Transport)
}

func TestTimeout(t *testing.T) {
	var cfg send.Config
	send.Timeout(1234 * time.Millisecond)(&cfg)
	assert.Equal(t, 1234*time.Millisecond, cfg.Timeout)
}
