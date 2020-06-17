package template

import (
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	stubTitleFn = func(val string) string {
		return strings.ToTitle(val)
	}
	stubLowerFn = func(val string) string {
		return strings.ToLower(val)
	}
	stubFuncMapA = FuncMap{
		"title": stubTitleFn,
	}
	stubFuncMapB = FuncMap{
		"lower": stubLowerFn,
	}
)

func TestNewConfig(t *testing.T) {
	cases := map[string]struct {
		opts     []Option
		expected Config
	}{
		"Use": {
			opts: []Option{
				Use("tpl1", "/path/to/tpl1.html"),
				Use("tpl2", "/path/to/tpl2.html"),
			},
			expected: Config{
				Templates: map[string]string{
					"tpl1": "/path/to/tpl1.html",
					"tpl2": "/path/to/tpl2.html",
				},
				Funcs: template.FuncMap{},
			},
		},
		"UseDir": {
			opts: []Option{
				UseDir("/path/to/tpls1", "/path/to/tpls2"),
				UseDir("/path/to/tpls3"),
			},
			expected: Config{
				Templates: map[string]string{},
				TemplateDirs: []string{
					"/path/to/tpls1",
					"/path/to/tpls2",
					"/path/to/tpls3",
				},
				Funcs: template.FuncMap{},
			},
		},
		"UseFuncs": {
			opts: []Option{
				UseFuncs(stubFuncMapA, stubFuncMapB),
			},
			expected: Config{
				Templates: map[string]string{},
				Funcs: template.FuncMap{
					"title": stubTitleFn,
					"lower": stubLowerFn,
				},
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := newConfig(tcase.opts...)

			assert.Equal(t, tcase.expected.Templates, cfg.Templates)
			assert.Equal(t, tcase.expected.TemplateDirs, cfg.TemplateDirs)

			for name := range tcase.expected.Funcs {
				_, ok := cfg.Funcs[name]
				assert.True(t, ok)
			}
		})
	}
}

func TestConfig_ParseTemplates(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	tplPath := func(path string) string {
		return filepath.Join(wd, "testdata", "templates", path)
	}

	tplDirPath := func(path string) string {
		return filepath.Join(wd, "testdata", "templateDirs", path)
	}

	cases := map[string]struct {
		opts         []Option
		expectedTpls []string
		expectedErr  error
	}{
		"only single templates": {
			opts: []Option{
				Use("tpl1", tplPath("tpl1.html")),
				Use("tpl2", tplPath("tpl2.html")),
				UseFuncs(stubFuncMapA),
			},
			expectedTpls: []string{"tpl1", "tpl2"},
			expectedErr:  nil,
		},
		"single template not found": {
			opts: []Option{
				Use("tpl1", tplPath("tpl10.html")),
				UseFuncs(stubFuncMapA),
			},
			expectedErr: &os.PathError{},
		},
		"template dirs": {
			opts: []Option{
				UseDir(tplDirPath("dir1"), tplDirPath("dir2")),
				UseFuncs(stubFuncMapA),
			},
			expectedTpls: []string{"tpl3", "tpl4", "tpl5", "tpl6", "nested.tpl7"},
		},
		"template dir not found": {
			opts: []Option{
				UseDir(tplDirPath("dirx")),
				UseFuncs(stubFuncMapA),
			},
			expectedErr: &os.PathError{},
		},
		"missing func": {
			opts: []Option{
				UseDir(tplDirPath("dir1")),
			},
			expectedErr: &template.Error{},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := newConfig(tcase.opts...)

			tpls, err := cfg.ParseTemplates()

			if tcase.expectedErr != nil {
				assert.True(t, errors.As(err, &tcase.expectedErr))
				return
			}

			for _, tplname := range tcase.expectedTpls {
				tpl := tpls.Lookup(tplname)
				assert.NotNil(t, tpl)
			}
		})
	}
}
