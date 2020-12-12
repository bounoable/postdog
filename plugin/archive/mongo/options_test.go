package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	var s Store
	Database("testdb")(&s)
	assert.Equal(t, "testdb", s.databaseName)
}

func TestCollection(t *testing.T) {
	var s Store
	Collection("testcol")(&s)
	assert.Equal(t, "testcol", s.collectionName)
}

func TestCreateIndexes(t *testing.T) {
	var s Store
	assert.Equal(t, false, s.wantIndexes)
	CreateIndexes(true)(&s)
	assert.Equal(t, true, s.wantIndexes)
	CreateIndexes(false)(&s)
	assert.Equal(t, false, s.wantIndexes)
}
