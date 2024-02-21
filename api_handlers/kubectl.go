package api_handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/umarkotak/go-kubeseal-gui/utils/render"
	"gopkg.in/yaml.v2"
)

type (
	SealParams struct {
		SecretName string `json:"secret_name"`
		YamlValue  string `json:"yaml_value"`
	}
)

func (h *handlers) GetKubectlContexts(w http.ResponseWriter, r *http.Request) {
	k8sContexts, err := kubectl.GetContexts(r.Context())
	if err != nil {
		render.Error(w, 400, err, "kubectl get contexts error")
		return
	}

	render.Response(w, k8sContexts)
}

func (h *handlers) GetKubectlSecrets(w http.ResponseWriter, r *http.Request) {
	aliasName := r.PathValue("alias_name")
	clusterName := config.Get().ClusterMap[aliasName].Name
	if clusterName == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "missing cluster name")
		return
	}

	err := kubectl.UseContext(r.Context(), clusterName)
	if err != nil {
		render.Error(w, 400, err, "kubectl use context error")
		return
	}

	secrets, err := kubectl.GetSecretsNameString(r.Context())
	if err != nil {
		render.Error(w, 400, err, "kubectl get secrets error")
		return
	}

	render.Response(w, secrets)
}

func (h *handlers) KubectlSecretRead(w http.ResponseWriter, r *http.Request) {
	aliasName := r.PathValue("alias_name")
	secretName := r.PathValue("secret_name")
	clusterName := config.Get().ClusterMap[aliasName].Name

	if clusterName == "" || secretName == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "missing cluster name or secret name")
		return
	}

	err := kubectl.UseContext(r.Context(), clusterName)
	if err != nil {
		render.Error(w, 400, err, "kubectl use context error")
		return
	}

	yamlData, err := h.kubectlGetSecretDecodedYaml(r.Context(), secretName)
	if err != nil {
		render.Error(w, 400, err, "kubectl get secret decoded error")
		return
	}

	render.ResponseRaw(w, yamlData)
}

func (h *handlers) KubectlSeal(w http.ResponseWriter, r *http.Request) {
	aliasName := r.PathValue("alias_name")
	clusterName := config.Get().ClusterMap[aliasName].Name

	if clusterName == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "missing cluster name or secret name")
		return
	}

	if config.Get().ControllerName == "" || config.Get().ControllerNamespace == "" {
		err := fmt.Errorf("missing configuration controller name and namespace")
		render.Error(w, 400, err, "please configure the controller name and controller namespace first")
		return
	}

	bodyByte, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	params := SealParams{}

	err = json.Unmarshal(bodyByte, &params)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 400, err, "")
		return
	}

	var kubectlSecret kubectl.Secret
	err = yaml.Unmarshal([]byte(params.YamlValue), &kubectlSecret)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	if kubectlSecret.Metadata.Name == "" {
		kubectlSecret.Metadata.Name = params.SecretName
	}

	kubectlSecret.EncodeBase64()

	secretSealedYaml, err := h.executeKubeseal(r.Context(), clusterName, kubectlSecret)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "execute kubeseal error")
		return
	}

	render.ResponseRaw(w, secretSealedYaml)
}

func (h *handlers) KubectlSecretDiff(w http.ResponseWriter, r *http.Request) {
	aliasName := r.PathValue("alias_name")
	clusterName := config.Get().ClusterMap[aliasName].Name

	if clusterName == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "missing cluster name or secret name")
		return
	}

	err := kubectl.UseContext(r.Context(), clusterName)
	if err != nil {
		render.Error(w, 400, err, "kubectl use context error")
		return
	}

	bodyByte, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	params := SealParams{}

	err = json.Unmarshal(bodyByte, &params)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 400, err, "")
		return
	}

	var kubectlSecret kubectl.Secret
	err = yaml.Unmarshal([]byte(params.YamlValue), &kubectlSecret)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	if kubectlSecret.Metadata.Name == "" {
		kubectlSecret.Metadata.Name = params.SecretName
	}

	diffResult := h.getKubeSecretDiff(r.Context(), kubectlSecret)

	render.Response(w, diffResult)
}
