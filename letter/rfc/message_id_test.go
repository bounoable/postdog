package rfc_test

import (
	"strings"
	"testing"

	"github.com/bounoable/postdog/letter/rfc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDGenerator_withoutDomain(t *testing.T) {
	gen := rfc.UUIDGenerator("")
	id := gen.GenerateID(rfc.Mail{})
	assert.IsType(t, "", id)
	assert.Len(t, id, 36+9+2+1) // UUID + "localhost" + "<" & ">" + "@"

	parts := strings.Split(id, "@")
	assert.Len(t, parts, 2)

	left := parts[0]
	right := parts[1]

	assert.Equal(t, byte('<'), left[0])

	uid, err := uuid.Parse(left[1:])
	assert.Nil(t, err)
	assert.Equal(t, uid.String(), left[1:])

	assert.Equal(t, "localhost>", right)
}

func TestUUIDGenerator_withDomain(t *testing.T) {
	gen := rfc.UUIDGenerator("foo")
	id := gen.GenerateID(rfc.Mail{})
	assert.Len(t, id, 36+3+2+1) // UUID + "foo" + "<" & ">" + "@"

	parts := strings.Split(id, "@")
	assert.Len(t, parts, 2)

	left := parts[0]
	right := parts[1]

	assert.Equal(t, byte('<'), left[0])

	uid, err := uuid.Parse(left[1:])
	assert.Nil(t, err)
	assert.Equal(t, uid.String(), left[1:])

	assert.Equal(t, "foo>", right)
}
