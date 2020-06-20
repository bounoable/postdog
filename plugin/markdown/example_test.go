package markdown_test

import (
	"context"
	"strings"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/autowire"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/markdown"
	gm "github.com/bounoable/postdog/plugin/markdown/goldmark"
	"github.com/yuin/goldmark"
)

func Example_basic() {
	po := postdog.New(
		postdog.WithPlugin(
			markdown.Plugin(
				gm.Converter(goldmark.New()), // use goldmark Markdown parser
				markdown.OverrideHTML(true),  // plugin options
			),
		),
	)

	err := po.Send(context.Background(), letter.Write(
		letter.Text("# Heading 1\n# Heading 2"), // The HTML body of the letter will be replaced with the generated HTML
	))

	_ = err
}

func Example_autowire() {
	config := `
plugins:
  - name: markdown
	use: goldmark
`

	cfg, err := autowire.Load(strings.NewReader(config))
	if err != nil {
		panic(err)
	}

	po, err := cfg.Office(context.Background())
	if err != nil {
		panic(err)
	}

	err = po.Send(context.Background(), letter.Write(
		letter.Text("# Heading 1\n# Heading 2"),
	))
}

func Example_disable() {
	po := postdog.New(
		postdog.WithPlugin(
			markdown.Plugin(gm.Converter(goldmark.New())),
		),
	)

	ctx := markdown.Disable(context.Background()) // disable Markdown conversion for this context

	po.Send(ctx, letter.Write(letter.Text("# Heading"))) // letter.HTML will stay empty
}
