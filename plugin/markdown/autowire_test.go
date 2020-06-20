package markdown_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/bounoable/postdog/plugin/markdown"
	"github.com/stretchr/testify/assert"
)

var testName = "test"

func init() {
	markdown.RegisterConverter(testName, markdown.ConverterFactoryFunc(
		func(_ map[string]interface{}) (markdown.Converter, error) {
			return nopConverter{}, nil
		}),
	)
}

func TestAutowirePlugin(t *testing.T) {
	cases := map[string]struct {
		cfg           map[string]interface{}
		expectedError error
	}{
		"unregistered converter": {
			cfg: map[string]interface{}{
				"use": "unknown",
			},
			expectedError: markdown.UnregisteredConverterError{
				Name: "unknown",
			},
		},
		"registered converter": {
			cfg: map[string]interface{}{
				"use": testName,
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			plugin, err := markdown.AutowirePlugin(ctx, tcase.cfg)
			assert.True(t, errors.Is(err, tcase.expectedError), err)

			if tcase.expectedError == nil {
				assert.NotNil(t, plugin)
			}
		})
	}
}

type nopConverter struct{}

func (conv nopConverter) Convert(_ []byte, _ io.Writer) error {
	return nil
}
