package template

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := newConfig(tcase.opts...)
			assert.Equal(t, tcase.expected, cfg)
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
			},
			expectedTpls: []string{"tpl1", "tpl2"},
			expectedErr:  nil,
		},
		"single template not found": {
			opts: []Option{
				Use("tpl1", tplPath("tpl10.html")),
			},
			expectedErr: &os.PathError{},
		},
		"template dirs": {
			opts: []Option{
				UseDir(tplDirPath("dir1"), tplDirPath("dir2")),
			},
			expectedTpls: []string{"tpl3", "tpl4", "tpl5", "tpl6", "nested.tpl7"},
		},
		"template dir not found": {
			opts: []Option{
				UseDir(tplDirPath("dirx")),
			},
			expectedErr: &os.PathError{},
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
