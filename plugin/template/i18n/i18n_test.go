package i18n_test

import (
	"fmt"
	"testing"

	"github.com/bounoable/postdog/plugin/template"
	"github.com/bounoable/postdog/plugin/template/i18n"
	"github.com/bounoable/postdog/plugin/template/i18n/mock_i18n"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trans := mock_i18n.NewMockTranslator(ctrl)
	opt := i18n.Use(trans)

	cfg := template.Config{Funcs: template.FuncMap{}}
	opt(&cfg)

	transFn, ok := cfg.Funcs["$t"].(i18n.TranslatorFunc)
	assert.True(t, ok)

	data := map[string]interface{}{
		"key1": "val1",
		"key2": 2,
	}

	trans.EXPECT().Translate("the.key", "en", data).DoAndReturn(func(key, _ string, data map[string]interface{}) (string, error) {
		return fmt.Sprintf("%s: %s %d", key, data["key1"], data["key2"]), nil
	})

	res, err := transFn("the.key", "en", data)
	assert.Nil(t, err)
	assert.Equal(t, "the.key: val1 2", res)
}
