# postdog - Template plugin

This plugin provides template support for letters.

## Example

```go
package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/template"
)

func main() {
  wd, _ := os.Getwd()

	po := postdog.New(
		postdog.WithPlugin(
			template.Plugin(
				template.UseDir( // add template directories
					filepath.Join(wd, "testdata/templateDirs/dir1"),
					filepath.Join(wd, "testdata/templateDirs/dir2"),
				),
				template.Use("custom1", filepath.Join(wd, "testdata/templates/tpl1.html")), // add single template
				template.Use("custom2", filepath.Join(wd, "testdata/templates/tpl2.html")),
				template.UseFuncs(template.FuncMap{ // register template functions
					"title": strings.Title,
					"upper": strings.ToUpper,
				}),
			),
		),
	)

	// Enable plugin for this context, use the template "dir2.nested.tpl7" and make custom data available in the template.
	ctx := template.Enable(context.Background(), "dir2.nested.tpl7", map[string]interface{}{
		"Name": "bob",
		"Now":  time.Now(),
	})

	err := po.Send(ctx, letter.Write(
		letter.HTML(`Hello {{ title .Data.Name }}, today is {{ .Data.Now.Format "2006/01/02" }}.`),
	))

	_ = err
}
```
