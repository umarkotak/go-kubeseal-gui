package api_handlers

import "github.com/umarkotak/go-kubeseal-gui/templates"

type handlers struct {
}

var (
	successTmpl, _ = templates.Get("success.html")
	failureTmpl, _ = templates.Get("failure.html")
)

func New() handlers {
	return handlers{}
}
