package handlers

import (
	"net/http"

	"github.com/umarkotak/go-kubeseal-gui/templates"
)

func (h *handlers) PageHome(w http.ResponseWriter, r *http.Request) {
	// tmpl := h.templateMap["home.html"]
	tmpl, _ := templates.Get("home.html", "base.html")

	tmpl.ExecuteTemplate(w, "base", nil)
}
