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

	w.Header().Add("HX-Refresh", "true")
	// successTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
	// 	"Message": "success add cluster!",
	// })
}

func (h *handlers) RemoveClustersConfig(w http.ResponseWriter, r *http.Request) {
	err := config.RemoveCluster(r.PathValue("alias"))
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "remove cluster config error",
		})
		return
	}

	successTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
		"Message": "success remove cluster config!",
	})
}

func (h *handlers) ClusterEnableSecrets(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	secrets := r.Form["secrets"]

	err := config.SetClusterSecret(r.PathValue("alias"), secrets)
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "enable secrets error",
		})
		return
	}

	w.Header().Add("HX-Refresh", "true")
}

func (h *handlers) SetupConfigController(w http.ResponseWriter, r *http.Request) {
	controllerName := r.FormValue("controller-name")
	controllerNamespace := r.FormValue("controller-namespace")

	err := config.SetController(controllerName, controllerNamespace)
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "save controller config error",
		})
		return
	}

	successTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
		"Message": "success setup controller config!",
	})
}

func (h *handlers) SetupGitIntegration(w http.ResponseWriter, r *http.Request) {
	err := config.SetGitIntConf(config.GitConf{
		GitProvider:       r.FormValue("git_conf_git_provider"),
		GitlabAccessToken: r.FormValue("git_conf_gitlab_access_token"),
		PrivateKeyPath:    r.FormValue("git_conf_private_key_path"),
		TmpFolderPath:     r.FormValue("git_conf_tmp_folder_path"),
		RepoUrl:           r.FormValue("git_conf_repo_url"),
		RepoHttpUrl:       r.FormValue("git_conf_repo_http_url"),
	})
	if err != nil {
		failureTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
			"Error":   err.Error(),
			"Message": "save git integration config error",
		})
		return
	}

	successTmpl.ExecuteTemplate(w, "notification", map[string]interface{}{
		"Message": "success setup git integration config!",
	})
}
