package page_handlers

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/umarkotak/go-kubeseal-gui/templates"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
)

func (h *handlers) Home(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := templates.Get("home.html", "base.html")

	clusterList := []string{}
	clusterSecretMap := map[string][]string{}

	for _, oneCluster := range config.Get().ClusterMap {
		clusterList = append(clusterList, oneCluster.Alias)

		for _, oneSecret := range oneCluster.RegisteredSecrets {
			clusterSecretMap[oneCluster.Alias] = append(clusterSecretMap[oneCluster.Alias], oneSecret.Name)
		}
	}

	slices.Sort(clusterList)

	clusterSecretMapByte, _ := json.Marshal(clusterSecretMap)

	tmpl.ExecuteTemplate(w, "base", struct {
		Clusters         []string
		ClusterSecretMap string
	}{
		Clusters:         clusterList,
		ClusterSecretMap: string(clusterSecretMapByte),
	})
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
		AddedClusters       []config.Cluster
		ControllerName      string
		ControllerNamespace string
	}{
		Clusters:            clusters,
		AddedClusters:       addedClusters,
		ControllerName:      config.Get().ControllerName,
		ControllerNamespace: config.Get().ControllerNamespace,
	})
}
