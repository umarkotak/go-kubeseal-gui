package api_handlers

import (
	"net/http"

	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
)

func (h *handlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	render.Response(w, config.Get())
}

func (h *handlers) AddClustersConfig(w http.ResponseWriter, r *http.Request) {
	k8sContext := r.FormValue("cluster-select")
	alias := r.FormValue("cluster-alias")

	err := kubectl.UseContext(r.Context(), k8sContext)
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "kubectl get context error",
		})
		return
	}

	allSecrets, err := kubectl.GetSecretsName(r.Context())
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "kubectl get secrets error",
		})
		return
	}

	cluster := config.Cluster{
		Alias:             alias,
		Name:              k8sContext,
		RegisteredSecrets: []config.Secret{},
		AllSecrets:        allSecrets,
	}

	err = config.SetCluster(cluster)
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "save cluster config error",
		})
		return
	}

	successTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
		"Message": "success add cluster!",
	})
}

func (h *handlers) SetupConfigController(w http.ResponseWriter, r *http.Request) {
	controllerName := r.FormValue("controller-name")
	controllerNamespace := r.FormValue("controller-namespace")

	err := config.SetController(controllerName, controllerNamespace)
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "save controller config",
		})
		return
	}

	successTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
		"Message": "success setup controller config!",
	})
}
