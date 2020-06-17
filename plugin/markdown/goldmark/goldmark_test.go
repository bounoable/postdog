package goldmark_test

import (
	"testing"

	"github.com/bounoable/postdog/plugin/markdown"
	gm "github.com/bounoable/postdog/plugin/markdown/goldmark"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
)

func TestConverter(t *testing.T) {
	var conv markdown.Converter = gm.Converter(goldmark.New())
	assert.NotNil(t, conv)
}