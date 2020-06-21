package memorystore_test

import (
	"testing"

	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/memorystore"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/bounoable/postdog/plugin/store/storetest"
)

func TestStore_Insert(t *testing.T) {
	storetest.Insert(t, memorystore.New())
}

func TestStore_Query(t *testing.T) {
	storetest.Query(t, func(letters ...store.Letter) query.Repository {
		return memorystore.New(letters...)
	})
}

func TestStore_Get(t *testing.T) {
	storetest.Get(t, func(letters ...store.Letter) query.Repository {
		return memorystore.New(letters...)
	})
}
