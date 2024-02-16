package api_handlers

import (
	"net/http"

	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
)

func (h *handlers) GetKubectlContexts(w http.ResponseWriter, r *http.Request) {
	k8sContexts, _ := kubectl.GetContexts(r.Context())

	render.Response(w, k8sContexts)
}
