package i18n

import "github.com/bounoable/postdog/plugin/template"

//go:generate mockgen -source=i18n.go -destination=./mock_i18n/i18n.go

// Translator translates message keys to human readable messages.
type Translator interface {
	Translate(key, locale string, data map[string]interface{}) (string, error)
}

// TranslatorFunc allows functions to be used as a `Translator`.
type TranslatorFunc func(string, string, map[string]interface{}) (string, error)

// Translate translates message keys to human readable messages.
func (fn TranslatorFunc) Translate(key, locale string, data map[string]interface{}) (string, error) {
	return fn(key, locale, data)
}

// Use registers the `$t` function in the templates. `$t` delegates the calls to `trans.Translate()`.
func Use(trans Translator) template.Option {
	return template.UseFuncs(template.FuncMap{
		"$t": TranslatorFunc(trans.Translate),
	})
}
