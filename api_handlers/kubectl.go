package api_handlers

import (
	"fmt"
	"net/http"

	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
	"gopkg.in/yaml.v2"
)

func (h *handlers) GetKubectlContexts(w http.ResponseWriter, r *http.Request) {
	k8sContexts, err := kubectl.GetContexts(r.Context())
	if err != nil {
		render.Error(w, 400, err, "kubectl get contexts error")
		return
	}

	render.Response(w, k8sContexts)
}

func (h *handlers) KubectlSecretRead(w http.ResponseWriter, r *http.Request) {
	aliasName := r.PathValue("alias_name")
	secretName := r.PathValue("secret_name")

	if aliasName == "" || secretName == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "missing cluster name or secret name")
		return
	}

	clusterName := config.Get().ClusterMap[aliasName].Name

	err := kubectl.UseContext(r.Context(), clusterName)
	if err != nil {
		render.Error(w, 400, err, "kubectl use context error")
		return
	}

	kubeSecretBase64, err := kubectl.GetSecretYaml(r.Context(), secretName)
	if err != nil {
		render.Error(w, 400, err, "kubectl get secret yaml error")
		return
	}

	err = kubeSecretBase64.DecodeBase64()
	if err != nil {
		render.Error(w, 400, err, "decode secret value error")
		return
	}

	yamlData, _ := yaml.Marshal(kubeSecretBase64)

	render.ResponseRaw(w, yamlData)
}
