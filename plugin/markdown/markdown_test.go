package markdown_test

import (
	"context"
	"io"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/mock_postdog"
	"github.com/bounoable/postdog/plugin/markdown"
	"github.com/bounoable/postdog/plugin/markdown/mock_markdown"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var mdStub string = `# Heading 1
## Heading 2`

var expectedHTML string = `<h1>Heading 1</h1><h2>Heading 2</h2>`

func TestPlugin(t *testing.T) {
	cases := map[string]struct {
		text               string
		html               string
		configureConverter func(*mock_markdown.MockConverter)
		opts               []markdown.Option
		expectedHTML       string
	}{
		"default": {
			text:         mdStub,
			expectedHTML: expectedHTML,
			configureConverter: func(conv *mock_markdown.MockConverter) {
				conv.EXPECT().Convert([]byte(mdStub), gomock.Any()).DoAndReturn(func(_ []byte, w io.Writer) error {
					w.Write([]byte(expectedHTML))
					return nil
				})
			},
		},
		"default (with html)": {
			text:         mdStub,
			html:         "filled",
			expectedHTML: "filled",
		},
		"override html": {
			text:         mdStub,
			html:         "filled",
			expectedHTML: expectedHTML,
			opts: []markdown.Option{
				markdown.OverrideHTML(true),
			},
			configureConverter: func(conv *mock_markdown.MockConverter) {
				conv.EXPECT().Convert([]byte(mdStub), gomock.Any()).DoAndReturn(func(_ []byte, w io.Writer) error {
					w.Write([]byte(expectedHTML))
					return nil
				})
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			let := letter.Write(letter.Content(tcase.text, tcase.html))

			converter := mock_markdown.NewMockConverter(ctrl)
			if tcase.configureConverter != nil {
				tcase.configureConverter(converter)
			}

			off := postdog.New(postdog.WithPlugin(markdown.Plugin(converter, tcase.opts...)))

			trans := mock_postdog.NewMockTransport(ctrl)
			trans.EXPECT().Send(ctx, gomock.Any()).Return(nil)
			off.ConfigureTransport("test", trans)

			err := off.Send(ctx, let)
			assert.Nil(t, err)
		})
	}
}

func TestDisable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	assert.False(t, markdown.Disabled(ctx))
	assert.True(t, markdown.Enabled(ctx))
	ctx = markdown.Disable(ctx)
	assert.True(t, markdown.Disabled(ctx))
	assert.False(t, markdown.Enabled(ctx))
	ctx = markdown.Enable(ctx)
	assert.False(t, markdown.Disabled(ctx))
	assert.True(t, markdown.Enabled(ctx))
	ctx = markdown.Disable(ctx)

	conv := mock_markdown.NewMockConverter(ctrl)

	off := postdog.New(postdog.WithPlugin(markdown.Plugin(conv)))
	trans := mock_postdog.NewMockTransport(ctrl)
	off.ConfigureTransport("test", trans)

	let := letter.Write(letter.Text("# Heading"))

	trans.EXPECT().Send(ctx, let).DoAndReturn(func(ctx context.Context, let letter.Letter) error {
		assert.Equal(t, "", let.HTML)
		return nil
	})

	off.Send(ctx, let)
}
