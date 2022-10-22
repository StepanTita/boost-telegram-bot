package config

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/pkg/errors"
)

type Templator interface {
	Template(name string) string
}

type templator struct {
	templates map[string]string
}

//go:embed templates/*
var embedTemplates embed.FS

// TODO: remove templates dir
func NewTemplator(templatesDir string) Templator {
	templates := make(map[string]string)
	err := fs.WalkDir(embedTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		// expect name.command.tmpl
		commandName := strings.Split(d.Name(), ".")[0]
		rawContent, err := fs.ReadFile(embedTemplates, path)
		if err != nil {
			return errors.Wrap(err, "failed to read file")
		}
		templates[commandName] = string(rawContent)

		return nil
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to walk dir"))
	}

	return &templator{
		templates: templates,
	}
}

func (l templator) Template(name string) string {
	return l.templates[name]
}
