package memory_test

import (
	"testing"

	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/memory"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/bounoable/postdog/plugin/store/storetest"
)

func TestStore_Insert(t *testing.T) {
	storetest.Insert(t, memory.NewStore())
}

func TestStore_Query(t *testing.T) {
	storetest.Query(t, func(letters ...store.Letter) query.Repository {
		return memory.NewStore(letters...)
	})
}
