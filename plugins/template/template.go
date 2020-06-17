// Package template provides template support for letter bodies.
package template

import (
	"context"
	"html/template"
	"strings"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
)

type plugin struct {
	cfg  Config
	tpls *template.Template
}

// Plugin creates the template plugin.
// It panics if it fails to parse the templates.
// Use TryPlugin() if you need to catch parse errors.
//
// Example:
//	plugin := template.Plugin(
//		template.UseDir("/templates")
//	)
func Plugin(opts ...Option) office.Plugin {
	plugin, err := TryPlugin(opts...)
	if err != nil {
		panic(err)
	}
	return plugin
}

// TryPlugin creates the template plugin. It doesn't panic when it fails to parse the templates.
func TryPlugin(opts ...Option) (office.Plugin, error) {
	cfg := newConfig(opts...)
	tpls, err := cfg.ParseTemplates()
	if err != nil {
		return nil, err
	}
	return plugin{
		cfg:  cfg,
		tpls: tpls,
	}, nil
}

func (p plugin) Install(pctx office.PluginContext) {
	pctx.WithMiddleware(
		office.MiddlewareFunc(func(
			ctx context.Context,
			let letter.Letter,
			next func(context.Context, letter.Letter) (letter.Letter, error),
		) (letter.Letter, error) {
			name, ok := Name(ctx)
			if !ok {
				return next(ctx, let)
			}

			var builder strings.Builder

			if err := p.tpls.ExecuteTemplate(&builder, name, struct {
				Letter letter.Letter
				Data   interface{}
			}{
				Letter: let,
				Data:   Data(ctx),
			}); err != nil {
				return let, err
			}

			let.HTML = builder.String()

			return next(ctx, let)
		}),
	)
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

// Name returns the template that should be used for sending letter with ctx.
// Returns false if ctx has no template set.
func Name(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(ctxTemplate).(string)
	return name, ok
}

// Data returns the data that is attached to ctx.
func Data(ctx context.Context) interface{} {
	return ctx.Value(ctxData)
}

type ctxKey string

var (
	ctxTemplate = ctxKey("template")
	ctxData     = ctxKey("data")
)
