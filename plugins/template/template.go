// Package template provides template support for letter bodies.
package template

import (
	"context"
	"strings"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

// Plugin creates the template plugin.
// It panics if it fails to parse the templates.
// Use TryPlugin() if you need to catch parse errors.
//
// Example:
//	plugin := template.Plugin(
//		template.UseDir("/templates")
//	)
func Plugin(opts ...Option) office.PluginFunc {
	plugin, err := TryPlugin(opts...)
	if err != nil {
		panic(err)
	}
	return plugin
}

// TryPlugin creates the template plugin. It doesn't panic when it fails to parse the templates.
func TryPlugin(opts ...Option) (office.PluginFunc, error) {
	cfg := newConfig(opts...)
	tpls, err := cfg.ParseTemplates()
	if err != nil {
		return nil, err
	}

	return func(pctx office.PluginContext) {
		pctx.WithMiddleware(
			office.MiddlewareFunc(func(
				ctx context.Context,
				let letter.Letter,
				next func(context.Context, letter.Letter) (letter.Letter, error),
			) (letter.Letter, error) {
				name, data, ok := For(ctx)
				if !ok {
					return next(ctx, let)
				}

				var builder strings.Builder

				if err := tpls.ExecuteTemplate(&builder, name, struct {
					Letter letter.Letter
					Data   interface{}
				}{
					Letter: let,
					Data:   data,
				}); err != nil {
					return let, err
				}

				let.HTML = builder.String()

				return next(ctx, let)
			}),
		)
	}, nil
}

// Enable sets the template that will be used to build the letter body for this context.
// Use the optional data that can be accesed in the template via {{ .Data }}
func Enable(ctx context.Context, name string, data interface{}) context.Context {
	ctx = context.WithValue(ctx, ctxTemplate, name)
	return context.WithValue(ctx, ctxData, data)
}

// Disable disables templates for this context.
func Disable(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, ctxTemplate, nil)
	return context.WithValue(ctx, ctxData, nil)
}

// Name returns the template that should be used for sending letters with ctx.
// Returns false if ctx has no template set.
func Name(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(ctxTemplate).(string)
	return name, ok
}

// Data returns the data that is attached to ctx.
func Data(ctx context.Context) interface{} {
	return ctx.Value(ctxData)
}

// For returns the template and data that should be used for sending letters with ctx.
// Returns false if ctx has no template set.
func For(ctx context.Context) (string, interface{}, bool) {
	name, ok := Name(ctx)
	return name, Data(ctx), ok
}

type ctxKey string

var (
	ctxTemplate = ctxKey("template")
	ctxData     = ctxKey("data")
)
