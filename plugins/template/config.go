package template

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config ...
type Config struct {
	Templates    map[string]string
	TemplateDirs []string
}

func newConfig(opts ...Option) Config {
	cfg := Config{
		Templates: map[string]string{},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

// Option ...
type Option func(*Config)

// Use ...
func Use(name, path string) Option {
	return func(cfg *Config) {
		cfg.Templates[name] = path
	}
}

// UseDir ...
func UseDir(dirs ...string) Option {
	return func(cfg *Config) {
		cfg.TemplateDirs = append(cfg.TemplateDirs, dirs...)
	}
}

// ParseTemplates ...
func (cfg Config) ParseTemplates() (*template.Template, error) {
	tpls := template.New("templates")

	for _, dir := range cfg.TemplateDirs {
		if err := extractTemplates(dir, tpls); err != nil {
			return nil, err
		}
	}

	for name, path := range cfg.Templates {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("could not open template '%s' at path '%s': %w", name, path, err)
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		f.Close()

		if _, err = tpls.New(name).Parse(string(b)); err != nil {
			return nil, fmt.Errorf("could not parse template '%s' at path '%s': %w", name, path, err)
		}
	}

	return tpls, nil
}

var slashExpr = regexp.MustCompile("^/")
var suffixExpr = regexp.MustCompile(`(?i)(\.[a-z0-9]+)+$`)

func extractTemplates(dir string, tpls *template.Template) error {
	dir = filepath.Clean(dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		name := strings.Replace(path, dir, "", 1)
		name = slashExpr.ReplaceAllString(name, "")
		name = suffixExpr.ReplaceAllString(name, "")
		name = strings.ReplaceAll(name, "/", ".")

		_, err = tpls.New(name).Parse(string(b))

		return err
	})
}
