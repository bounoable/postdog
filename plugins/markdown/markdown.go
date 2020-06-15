// Package markdown provides Markdown support for letters.
// It registers a middleware that converts the text body of letters with a Markdown converter
// and sets the HTML body to the conversion result.
package markdown

//go:generate mockgen -source=markdown.go -destination=./mock_markdown/markdown.go

import (
	"bytes"
	"context"
	"io"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

// Converter converts a Markdown source to HTML.
type Converter interface {
	Convert(src []byte, w io.Writer) error
}

// Config is the plugin configuration.
type Config struct {
	// Override HTML field even if it's already filled.
	OverrideHTML bool
}

// Plugin is the install function for the Markdown plugin.
// It takes the Text field of the outgoing letters, converts them and sets the HTML field to the result.
func Plugin(conv Converter, opts ...Option) office.PluginFunc {
	return PluginWithConfig(conv, newConfig(opts...))
}

// PluginWithConfig is the install function for the Markdown plugin.
// It takes the Text field of the outgoing letters, converts them and sets the HTML field to the result.
func PluginWithConfig(conv Converter, cfg Config) office.PluginFunc {
	return func(pctx office.PluginContext) {
		pctx.WithMiddleware(
			office.MiddlewareFunc(func(_ context.Context, let letter.Letter) (letter.Letter, error) {
				if len(let.HTML) > 0 && !cfg.OverrideHTML {
					return let, nil
				}

				var buf bytes.Buffer
				if err := conv.Convert([]byte(let.Text), &buf); err != nil {
					return let, err
				}
				let.HTML = buf.String()
				return let, nil
			}),
		)
	}
}

func newConfig(opts ...Option) Config {
	var cfg Config
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// Option is a plugin option.
type Option func(*Config)

// OverrideHTML overrides the HTML body of the letter, even if it is already filled.
func OverrideHTML(override bool) Option {
	return func(cfg *Config) {
		cfg.OverrideHTML = override
	}
}
