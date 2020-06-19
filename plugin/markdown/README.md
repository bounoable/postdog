# postdog - Markdown Plugin

This plugin provides Markdown support for mails.

## Example

```go
package main

import (
  "context"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/markdown"
	gm "github.com/bounoable/postdog/plugin/markdown/goldmark"
	"github.com/yuin/goldmark"
)

func main() {
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
}
```
