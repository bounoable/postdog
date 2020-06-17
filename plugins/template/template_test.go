package template_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/office"
	"github.com/bounoable/postdog/office/mock_office"
	"github.com/bounoable/postdog/plugins/template"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEnable(t *testing.T) {
	ctx := context.Background()

	name, ok := template.Name(ctx)
	assert.Empty(t, name)
	assert.False(t, ok)

	ctx = template.Enable(ctx, "test")
	name, ok = template.Name(ctx)
	assert.Equal(t, "test", name)
	assert.True(t, ok)

	ctx = template.Disable(ctx)
	name, ok = template.Name(ctx)
	assert.Empty(t, name)
	assert.False(t, ok)
}

func TestPlugin(t *testing.T) {
	wd, _ := os.Getwd()

	plugin := template.Plugin(
		template.Use("test1", filepath.Join(wd, "testdata", "templates", "tpl1.html")),
		template.Use("test2", filepath.Join(wd, "testdata", "templates", "tpl2.html")),
		template.UseDir(filepath.Join(wd, "testdata", "templateDirs", "dir1")),
		template.UseDir(filepath.Join(wd, "testdata", "templateDirs", "dir2")),
	)

	names := []string{
		"test1",
		"test2",
		"tpl3",
		"tpl4",
		"tpl5",
		"tpl6",
		"nested.tpl7",
	}

	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			ctx = template.Enable(ctx, name)

			off := office.New(office.WithPlugin(plugin))

			let := letter.Write(letter.Text("example body"))
			expectedLet := makeExpectedLetter(let, i+1)

			trans := mock_office.NewMockTransport(ctrl)
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
  example body
</p>
`, num)
	return let
}
