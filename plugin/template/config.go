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

// Config is the plugin configuration.
type Config struct {
	Templates    map[string]string
	TemplateDirs []DirectoryConfig
	Funcs        FuncMap
}

func newConfig(opts ...Option) Config {
	cfg := Config{
		Templates: map[string]string{},
		Funcs:     FuncMap{},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

// Option is a plugin option.
type Option func(*Config)

// Use registers the template at filepath under name.
func Use(name, filepath string) Option {
	return func(cfg *Config) {
		cfg.Templates[name] = filepath
	}
}

// UseDir registers all files in dirs and their subdirectories as templates.
// The template name will be set to the relative path of the file to the given directory in dirs,
// where every directory separator is replaced by a dot and the file extensions is removed.
//
// Example:
//	Given the following files:
// 	/templates/tpl1.html
//	/templates/tpl2.html
//	/templates/nested/tpl3.html
//	/templates/nested/deeper/tpl4.html
//
//	UseDir("/templates") will result in the following template names:
//	- tpl1
//	- tpl2
//	- nested.tpl3
//	- nested.deeper.tpl4
//
// Use the `Exclude()` option to filter templates based on their filepath.
func UseDir(dir string, opts ...UseDirOption) Option {
	dirCfg := DirectoryConfig{Dir: dir}
	for _, opt := range opts {
		opt(&dirCfg)
	}
	return func(cfg *Config) {
		cfg.TemplateDirs = append(cfg.TemplateDirs, dirCfg)
	}
}

// UseDirOption is an option for the `UseDir()` function.
type UseDirOption func(*DirectoryConfig)

// Exclude filters templates by their filepath.
func Exclude(fn func(string) bool) UseDirOption {
	return func(cfg *DirectoryConfig) {
		cfg.Exclude = fn
	}
}

// ExcludePattern filters templates by a pattern.
// This function panics if it can't compile the pattern.
func ExcludePattern(pattern string) UseDirOption {
	return ExcludeRegex(regexp.MustCompile(pattern))
}

// ExcludeRegex filters templates by a regular expression.
func ExcludeRegex(expr *regexp.Regexp) UseDirOption {
	return Exclude(func(path string) bool {
		return expr.MatchString(path)
	})
}

// DirectoryConfig is the configuration for a template directory.
type DirectoryConfig struct {
	Dir     string
	Exclude func(string) bool
}

// UseFuncs adds the functions in funcMaps the templates' function maps.
func UseFuncs(funcMaps ...FuncMap) Option {
	return func(cfg *Config) {
		for _, funcs := range funcMaps {
			for name, fn := range funcs {
				cfg.Funcs[name] = fn
			}
		}
	}
}

// ParseTemplates parses the templates that are configured in cfg and returns the root template.
func (cfg Config) ParseTemplates() (*template.Template, error) {
	tpls := template.New("templates").Funcs(template.FuncMap(cfg.Funcs))

	for _, dirCfg := range cfg.TemplateDirs {
		if err := parseTemplates(dirCfg, tpls); err != nil {
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

var slashExpr = regexp.MustCompile(fmt.Sprintf("^%c", os.PathSeparator))
var suffixExpr = regexp.MustCompile(`(?i)(\.[a-z0-9]+)+$`)

func parseTemplates(dirCfg DirectoryConfig, tpls *template.Template) error {
	dir := filepath.Clean(dirCfg.Dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if dirCfg.Exclude != nil && dirCfg.Exclude(path) {
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
		name = strings.ReplaceAll(name, string(os.PathSeparator), ".")

		_, err = tpls.New(name).Parse(string(b))

		return err
	})
}
