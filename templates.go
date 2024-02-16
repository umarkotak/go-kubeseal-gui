package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	// go:embed templates/*.html
	files embed.FS

	templatesMap = map[string]*template.Template{}
)

func LoadTemplatesEmbeded() (map[string]*template.Template, error) {
	tmplFiles, err := fs.ReadDir(files, "templates")
	if err != nil {
		logrus.Error(err)
		return templatesMap, err
	}

	for _, tmpl := range tmplFiles {
		pt, err := template.ParseFS(files, fmt.Sprintf("templates/%s", tmpl.Name()))
		if err != nil {
			logrus.Error(err)
			return templatesMap, err
		}

		templatesMap[tmpl.Name()] = pt
	}

	return templatesMap, nil
}

func LoadTemplates() (map[string]*template.Template, error) {
	tmplFiles, err := os.ReadDir("templates")
	if err != nil {
		logrus.Error(err)
		return templatesMap, err
	}

	for _, tmpl := range tmplFiles {
		pt, err := template.ParseFiles(fmt.Sprintf("templates/%s", tmpl.Name()))
		if err != nil {
			logrus.Errorf("%v %v", err, tmpl.Name())
			return templatesMap, err
		}

		templatesMap[tmpl.Name()] = pt
	}

	return templatesMap, nil
}
