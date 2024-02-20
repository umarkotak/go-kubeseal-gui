package api_handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/helper"
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

func (h *handlers) KubectlSeal(w http.ResponseWriter, r *http.Request) {
	aliasName := r.PathValue("alias_name")
	clusterName := config.Get().ClusterMap[aliasName].Name

	if clusterName == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "missing cluster name or secret name")
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

	newSecretYaml, err := yaml.Marshal(&kubectlSecret)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	tempSecretYamlFileName := fmt.Sprintf("go-kubeseal-gui-temp-secret-%s-%s.yaml", clusterName, kubectlSecret.Metadata.Name)

	err = os.WriteFile(tempSecretYamlFileName, newSecretYaml, 0644)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	if config.Get().ControllerName == "" || config.Get().ControllerNamespace == "" {
		err = fmt.Errorf("missing configuration controller name and namespace")
		render.Error(w, 400, err, "please configure the controller name and controller namespace first")
		return
	}

	err = kubectl.UseContext(r.Context(), clusterName)
	if err != nil {
		render.Error(w, 400, err, "kubectl use context error")
		return
	}

	cmd := exec.Command(
		"kubeseal",
		fmt.Sprintf("--controller-name=%s", config.Get().ControllerName),
		fmt.Sprintf("--controller-namespace=%s", config.Get().ControllerNamespace),
		"--format=yaml",
		"-f",
		tempSecretYamlFileName,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("%s - %s", err.Error(), stderr.String())
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	tempSecretSealedYamlFileName := fmt.Sprintf("go-kubeseal-gui-temp-secret-sealed-%s-%s.yaml", clusterName, kubectlSecret.Metadata.Name)

	err = exec.Command("sh", "-c", fmt.Sprintf("echo '%v' > %v", string(output), tempSecretSealedYamlFileName)).Run()
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	secretSealedYaml, err := os.ReadFile(tempSecretSealedYamlFileName)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	err = exec.Command("rm", tempSecretYamlFileName).Run()
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	err = exec.Command("rm", tempSecretSealedYamlFileName).Run()
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
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

	oldKubectlSecret, err := kubectl.GetSecretYaml(r.Context(), kubectlSecret.Metadata.Name)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
	}

	err = oldKubectlSecret.DecodeBase64()
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
	}

	diffResult := helper.DiffMaps(oldKubectlSecret.Data, kubectlSecret.Data)

	render.Response(w, diffResult)
}
