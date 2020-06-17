// Package markdown provides Markdown support for letters.
// It registers a middleware that converts the text body of letters with a Markdown converter
// and sets the HTML body to the conversion result.
package markdown

//go:generate mockgen -source=markdown.go -destination=./mock_markdown/markdown.go

import (
	"bytes"
	"context"
	"io"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
)

// Converter converts a Markdown source to HTML.
type Converter interface {
	Convert(src []byte, w io.Writer) error
}

// ConverterFunc allows a function to be used as Converter.
type ConverterFunc func([]byte, io.Writer) error

// Convert converts the Markdown in src to HTML and writes the result to w.
func (fn ConverterFunc) Convert(src []byte, w io.Writer) error {
	return fn(src, w)
}

// Config is the plugin configuration.
type Config struct {
	// Override HTML field even if it's already filled.
	OverrideHTML bool
}

// Plugin is the install function for the Markdown plugin.
// It takes the Text field of the outgoing letters, converts them and sets the HTML field to the result.
func Plugin(conv Converter, opts ...Option) postdog.PluginFunc {
	return PluginWithConfig(conv, newConfig(opts...))
}

// PluginWithConfig is the install function for the Markdown plugin.
// It takes the Text field of the outgoing letters, converts them and sets the HTML field to the result.
func PluginWithConfig(conv Converter, cfg Config) postdog.PluginFunc {
	return func(pctx postdog.PluginContext) {
		pctx.WithMiddleware(
			postdog.MiddlewareFunc(func(
				ctx context.Context,
				let letter.Letter,
				next func(context.Context, letter.Letter) (letter.Letter, error),
			) (letter.Letter, error) {
				if Disabled(ctx) {
					return let, nil
				}

				if len(let.HTML) > 0 && !cfg.OverrideHTML {
					return let, nil
				}

				var buf bytes.Buffer
				if err := conv.Convert([]byte(let.Text), &buf); err != nil {
					return let, err
				}
				let.HTML = buf.String()

				return next(ctx, let)
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

// Disabled determines if Markdown conversion is disabled for ctx.
func Disabled(ctx context.Context) bool {
	disabled, _ := ctx.Value(ctxDisabled).(bool)
	return disabled
}

// Disable disables Markdown conversions for all letters that are sent with this ctx.
func Disable(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxDisabled, true)
}

// Enabled determines if Markdown conversion is enabled for ctx.
func Enabled(ctx context.Context) bool {
	return !Disabled(ctx)
}

// Enable (re)enables Markdown conversions for ctx.
func Enable(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxDisabled, false)
}

type ctxKey string

var ctxDisabled = ctxKey("disabled")
