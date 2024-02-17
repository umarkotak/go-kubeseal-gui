package page_handlers

import (
	"net/http"

	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/umarkotak/go-kubeseal-gui/templates"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
)

func (h *handlers) Home(w http.ResponseWriter, r *http.Request) {
	// tmpl := h.templateMap["home.html"]
	tmpl, _ := templates.Get("home.html", "base.html")

	tmpl.ExecuteTemplate(w, "base", nil)
}

func (h *handlers) Config(w http.ResponseWriter, r *http.Request) {
	clusters, err := kubectl.GetContexts(r.Context())
	if err != nil {
		render.Error(w, 500, err, "kubectl get contexts error")
		return
	}

	tmpl, _ := templates.Get("config.html", "base.html")

	addedClusters := []config.Cluster{}
	for _, oneCluster := range config.Get().ClusterMap {
		addedClusters = append(addedClusters, oneCluster)
	}

	tmpl.ExecuteTemplate(w, "base", struct {
		Clusters            []string
		ControllerName      string
		ControllerNamespace string
		AddedClusters       []config.Cluster
	}{
		ControllerName:      config.Get().ControllerName,
		ControllerNamespace: config.Get().ControllerNamespace,
		Clusters:            clusters,
		AddedClusters:       addedClusters,
	})
}
