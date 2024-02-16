package api_handlers

import (
	"net/http"

	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
)

func (h *handlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	render.Response(w, config.Get())
}
