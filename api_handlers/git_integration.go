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
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
)

type (
	SealAndPushParams struct {
		SecretName string `json:"secret_name"`
		YamlValue  string `json:"yaml_value"`

		Tag     string `json:"tag"`     // used to differentiate purpose
		Remarks string `json:"remarks"` // used as additional reference, eg: author name
	}
)

func (h *handlers) KubectlSealAndPush(w http.ResponseWriter, r *http.Request) {
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

	params := SealAndPushParams{}

	err = json.Unmarshal(bodyByte, &params)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 400, err, "")
		return
	}

	if params.Tag == "" || params.Remarks == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "params and remarks cannot empty")
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

	var kubectlSecretSealed kubectl.SecretSealed

	yaml.Unmarshal(secretSealedYaml, &kubectlSecretSealed)

	pushResponse, err := h.PushToGit(r.Context(), aliasName, kubectlSecret, kubectlSecretSealed, secretSealedYaml, params)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "execute push to git error")
		return
	}

	render.Response(w, pushResponse)
}

func (h *handlers) KubectlSealAndPr(w http.ResponseWriter, r *http.Request) {
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

	params := SealAndPushParams{}

	err = json.Unmarshal(bodyByte, &params)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 400, err, "")
		return
	}

	if params.Tag == "" || params.Remarks == "" {
		err := fmt.Errorf("bad request")
		render.Error(w, 400, err, "params and remarks cannot empty")
		return
	}

	if config.Get().GitConf.GitProvider == "gitlab" {
		if config.Get().GitConf.GitlabAccessToken == "" || config.Get().GitConf.GitlabBaseUrl == "" {
			err := fmt.Errorf("bad request")
			render.Error(w, 400, err, "gitlab access token and base url cannot empty")
			return
		}
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

	var kubectlSecretSealed kubectl.SecretSealed

	yaml.Unmarshal(secretSealedYaml, &kubectlSecretSealed)

	pushResponse, err := h.PushToGit(r.Context(), aliasName, kubectlSecret, kubectlSecretSealed, secretSealedYaml, params)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "execute push to git error")
		return
	}

	gitlabClient, err := gitlab.NewClient(
		config.Get().GitConf.GitlabAccessToken,
		gitlab.WithBaseURL(config.Get().GitConf.GitlabBaseUrl),
	)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "")
		return
	}

	masterBranchName := config.Get().GitConf.MasterBranchName
	if masterBranchName == "" {
		masterBranchName = "master"
	}

	mrParams := &gitlab.CreateMergeRequestOptions{
		SourceBranch: gitlab.Ptr(pushResponse.BranchName),
		TargetBranch: gitlab.Ptr(masterBranchName),
		Title:        gitlab.Ptr(fmt.Sprintf("merge request from %s to master", pushResponse.BranchName)),
	}

	mr, _, err := gitlabClient.MergeRequests.CreateMergeRequest(config.Get().GitConf.RepoEnvProjectID, mrParams)
	if err != nil {
		logrus.WithContext(r.Context()).Error(err)
		render.Error(w, 500, err, "create merge request")
		return
	}

	mrResponse := PRResponse{
		MrUrl:            fmt.Sprintf("%s/-/merge_requests/%v", config.Get().GitConf.RepoHttpUrl, mr.IID),
		SecretSealedYaml: pushResponse.SecretSealedYaml,
	}

	render.Response(w, mrResponse)
}
