package handlers

import "text/template"

type handlers struct {
	templateMap map[string]*template.Template
}

func New(templateMap map[string]*template.Template) handlers {
	return handlers{
		templateMap: templateMap,
	}
}
