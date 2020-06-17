package markdown_test

import (
	"context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/markdown"
	gm "github.com/bounoable/postdog/plugin/markdown/goldmark"
	"github.com/yuin/goldmark"
)

func Example_basic() {
	off := postdog.New(
		postdog.WithPlugin(
			markdown.Plugin(
				gm.Converter(goldmark.New()), // use goldmark Markdown parser
				markdown.OverrideHTML(true),  // plugin options
			),
		),
	)

	off.Send(context.Background(), letter.Write(letter.Text("# Heading"))) // letter.HTML will be set to <h1>Heading</h1>
}

func Example_disable() {
	off := postdog.New(
		postdog.WithPlugin(
			markdown.Plugin(gm.Converter(goldmark.New())),
		),
	)

	ctx := markdown.Disable(context.Background()) // disable Markdown conversion for this context

	off.Send(ctx, letter.Write(letter.Text("# Heading"))) // letter.HTML will stay empty
}
