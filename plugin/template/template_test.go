package template_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/mock_postdog"
	"github.com/bounoable/postdog/plugin/template"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEnable(t *testing.T) {
	ctx := context.Background()

	name, data, ok := template.For(ctx)
	assert.Empty(t, name)
	assert.Nil(t, data)
	assert.False(t, ok)

	ctx = template.Enable(ctx, "test", map[string]interface{}{
		"A": true,
		"B": 1,
		"C": "test",
	})

	name, data, ok = template.For(ctx)
	assert.Equal(t, "test", name)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{
		"A": true,
		"B": 1,
		"C": "test",
	}, data)

	ctx = template.Disable(ctx)
	name, data, ok = template.For(ctx)
	assert.Empty(t, name)
	assert.False(t, ok)
	assert.Nil(t, data)
}

func TestPlugin(t *testing.T) {
	wd, _ := os.Getwd()

	plugin := template.Plugin(
		template.Use("test1", filepath.Join(wd, "testdata", "templates", "tpl1.html")),
		template.Use("test2", filepath.Join(wd, "testdata", "templates", "tpl2.html")),
		template.UseDir(filepath.Join(wd, "testdata", "templateDirs", "dir1")),
		template.UseDir(filepath.Join(wd, "testdata", "templateDirs", "dir2"), template.Exclude(func(path string) bool {
			return strings.Contains(path, "tpl6")
		})),
		template.UseFuncs(template.FuncMap{
			"title": func(val string) string {
				return strings.Title(val)
			},
		}),
	)

	cases := []struct {
		name        string
		shouldExist bool
	}{
		{name: "test1", shouldExist: true},
		{name: "test2", shouldExist: true},
		{name: "tpl3", shouldExist: true},
		{name: "tpl4", shouldExist: true},
		{name: "tpl5", shouldExist: true},
		{name: "tpl6", shouldExist: false},
		{name: "nested.tpl7", shouldExist: true},
	}

	for i, tcase := range cases {
		t.Run(tcase.name, func(t *testing.T) {
			if !tcase.shouldExist {
				return
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			ctx = template.Enable(ctx, tcase.name, nil)

			off := postdog.New(postdog.WithPlugin(plugin))

			let := letter.Write(letter.Text("example body"))
			expectedLet := makeExpectedLetter(let, i+1)

			trans := mock_postdog.NewMockTransport(ctrl)
			trans.EXPECT().Send(ctx, expectedLet)

			off.ConfigureTransport("test", trans)

			err := off.Send(ctx, let)
			assert.Nil(t, err)
		})
	}
}

func makeExpectedLetter(let letter.Letter, num int) letter.Letter {
	let.HTML = fmt.Sprintf(`<h1>Template %d</h1>

<p>
  Example Body
</p>
`, num)
	return let
}
