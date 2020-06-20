// Package goldmark provides an adapter for the goldmark Markdown parser.
package goldmark

import (
	"io"

	"github.com/bounoable/postdog/plugin/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

// Name is the name of the converter.
var Name = "goldmark"

func init() {
	markdown.RegisterConverter(Name, markdown.ConverterFactoryFunc(
		func(_ map[string]interface{}) (markdown.Converter, error) {
			return Converter(goldmark.New()), nil
		}),
	)
}

// Converter creates a markdown.Converter from md.
func Converter(md goldmark.Markdown, opts ...Option) markdown.Converter {
	a := adapter{md: md}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

// Option is a goldmark adapter option.
type Option func(*adapter)

// ParseWith configures goldmark to use opts as parse options.
func ParseWith(opts ...parser.ParseOption) Option {
	return func(a *adapter) {
		a.parseOpts = append(a.parseOpts, opts...)
	}
}

type adapter struct {
	md        goldmark.Markdown
	parseOpts []parser.ParseOption
}

func (a adapter) Convert(src []byte, w io.Writer) error {
	return a.md.Convert(src, w, a.parseOpts...)
}
