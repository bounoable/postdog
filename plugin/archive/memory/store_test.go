package memory_test

import (
	"testing"

	"github.com/bounoable/postdog/plugin/archive"
	"github.com/bounoable/postdog/plugin/archive/memory"
	"github.com/bounoable/postdog/plugin/archive/test"
)

func TestStore(t *testing.T) {
	test.Store(t, func() archive.Store { return memory.NewStore() })
}
