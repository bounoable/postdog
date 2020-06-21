package template_test

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

func Example() {
	wd, _ := os.Getwd()

	po := postdog.New(
		postdog.WithPlugin(
			template.Plugin(
				// add template directories
				template.UseDir(filepath.Join(wd, "testdata/templateDirs/dir1")),
				template.UseDir(filepath.Join(wd, "testdata/templateDirs/dir2")),

				// add single templates
				template.Use("custom1", filepath.Join(wd, "testdata/templates/tpl1.html")),
				template.Use("custom2", filepath.Join(wd, "testdata/templates/tpl2.html")),

				// add template funcs
				template.UseFuncs(template.FuncMap{
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
