package api_handlers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/helper"
	"github.com/umarkotak/go-kubeseal-gui/kubectl"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
)

const (
	tempSecretFileName = "secret-sealed.yaml"
	pushAuthorName     = "go-kubeseal-gui"
	pushAuthorEmail    = "go-kubeseal-gui@dummy-email.com"
)

type (
	PushResponse struct {
		GitUrl           string `json:"git_url"`
		BranchUrl        string `json:"branch_url"`
		MergeRequestUrl  string `json:"merge_request_url"`
		SecretSealedYaml string `json:"secret_sealed_yaml"`
	}
)

func (h *handlers) kubectlGetSecretDecodedYaml(ctx context.Context, secretName string) ([]byte, error) {
	kubeSecretBase64, err := kubectl.GetSecretYaml(ctx, secretName)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}

	err = kubeSecretBase64.DecodeBase64()
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}

	yamlData, _ := yaml.Marshal(kubeSecretBase64)

	return yamlData, nil
}

// kubectlSecret data value should be encoded in base64;
// this will return yaml file result of kubeseal
func (h *handlers) executeKubeseal(ctx context.Context, clusterName string, kubectlSecret kubectl.Secret) ([]byte, error) {
	newSecretYaml, err := yaml.Marshal(kubectlSecret)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}

	tempSecretYamlFileName := fmt.Sprintf("go-kubeseal-gui-temp-secret-%s-%s.yaml", clusterName, kubectlSecret.Metadata.Name)

	err = os.WriteFile(tempSecretYamlFileName, newSecretYaml, 0644)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}
	defer func() {
		err = exec.Command("rm", tempSecretYamlFileName).Run()
		if err != nil {
			logrus.WithContext(ctx).Error(err)
		}
	}()

	err = kubectl.UseContext(ctx, clusterName)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
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
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}

	tempSecretSealedYamlFileName := fmt.Sprintf("go-kubeseal-gui-temp-secret-sealed-%s-%s.yaml", clusterName, kubectlSecret.Metadata.Name)

	err = exec.Command("sh", "-c", fmt.Sprintf("echo '%v' > %v", string(output), tempSecretSealedYamlFileName)).Run()
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}
	defer func() {
		err = exec.Command("rm", tempSecretSealedYamlFileName).Run()
		if err != nil {
			logrus.WithContext(ctx).Error(err)
		}
	}()

	secretSealedYaml, err := os.ReadFile(tempSecretSealedYamlFileName)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return []byte{}, err
	}

	return secretSealedYaml, nil
}

func (h *handlers) getKubeSecretDiff(ctx context.Context, kubectlSecret kubectl.Secret) map[string]map[string]string {
	oldKubectlSecret, err := kubectl.GetSecretYaml(ctx, kubectlSecret.Metadata.Name)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
	}

	err = oldKubectlSecret.DecodeBase64()
	if err != nil {
		logrus.WithContext(ctx).Error(err)
	}

	diffResult := helper.DiffMaps(oldKubectlSecret.Data, kubectlSecret.Data)

	return diffResult
}

func (h *handlers) PushToGit(
	ctx context.Context,
	aliasName string,
	kubectlSecretEncoded kubectl.Secret,
	kubectlSecretSealed kubectl.SecretSealed,
	secretSealedYaml []byte,
	params SealAndPushParams,
) (PushResponse, error) {

	auth, err := ssh.NewPublicKeysFromFile("git", config.Get().GitConf.PrivateKeyPath, "")
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return PushResponse{}, err
	}

	// Clean up tmp folder
	exec.Command("rm", "-rf", config.Get().GitConf.TmpFolderPath).Run()

	// Clone the repository to be pushed later
	_, err = git.PlainClone(config.Get().GitConf.TmpFolderPath, false, &git.CloneOptions{
		URL:      config.Get().GitConf.RepoUrl,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil && !strings.Contains(err.Error(), "repository already exists") {
		logrus.WithContext(ctx).Error(err)
		return PushResponse{}, err
	}

	// Open repository
	repo, err := git.PlainOpen(config.Get().GitConf.TmpFolderPath)
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	// result sample: tmp/go-kube-seal-temp-git/gcp-k8s-integration/identity-service/integration-1
	filePath := fmt.Sprintf(
		"%s/%s/%s/%s", config.Get().GitConf.TmpFolderPath, aliasName, kubectlSecretEncoded.Metadata.Name, params.Tag,
	)

	fullPath := fmt.Sprintf("%s/%s", filePath, tempSecretFileName)

	branchName := fmt.Sprintf("%s-%v", kubectlSecretEncoded.Metadata.Name, time.Now().Unix())

	branchRef := plumbing.NewBranchReferenceName(branchName)

	worktree, err := repo.Worktree()
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Create: true,
		Keep:   true,
	})
	if err != nil {
		logrus.Errorf("%s - %s", err.Error(), branchRef.Short())
		return PushResponse{}, err
	}

	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	updatingValue := map[string]string{}

	tmpKubectlSecret := kubectlSecretEncoded
	tmpKubectlSecret.DecodeBase64()

	diffResult := h.getKubeSecretDiff(ctx, tmpKubectlSecret)

	for k, v := range kubectlSecretSealed.Spec.EncryptedData {
		_, found := diffResult[k]
		if found {
			updatingValue[k] = v
		}
	}

	oldSecretSealedYaml, err := os.ReadFile(fullPath)
	if err == nil {
		var oldSecretSealed kubectl.SecretSealed
		err = yaml.Unmarshal(oldSecretSealedYaml, &oldSecretSealed)
		if err != nil {
			logrus.Error(err)
			return PushResponse{}, err
		}

		for k, v := range updatingValue {
			oldSecretSealed.Spec.EncryptedData[k] = v
		}

		oldSecretSealedYaml, err := yaml.Marshal(oldSecretSealed)
		if err != nil {
			logrus.Error(err)
			return PushResponse{}, err
		}

		secretSealedYaml = oldSecretSealedYaml
	}

	err = os.WriteFile(fullPath, secretSealedYaml, 0644)
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	_, err = worktree.Add(".")
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	_, err = worktree.Commit(
		fmt.Sprintf("%s %v", branchName, params.Remarks),
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  pushAuthorName,
				Email: pushAuthorEmail,
				When:  time.Now(),
			},
		})
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		logrus.Error("Error getting remote:", err)
		return PushResponse{}, err
	}

	err = remote.Push(&git.PushOptions{
		Auth:  auth,
		Force: true,
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec("+" + branchRef + ":" + branchRef),
		},
	})
	if err != nil {
		logrus.Error(err)
		return PushResponse{}, err
	}

	_, err = gitlab.NewClient(config.Get().GitConf.GitlabAccessToken)
	if err == nil {
		// TODO: create MR to gitlab
		// - https://github.com/xanzy/go-gitlab
		// - https://github.com/xanzy/go-gitlab/blob/main/merge_requests.go
		// - https://docs.gitlab.com/ee/api/merge_requests.html#create-mr
		// - https://docs.gitlab.com/ee/api/rest/index.html#namespaced-path-encoding
	}

	return PushResponse{
		GitUrl:           fmt.Sprintf("%s", config.Get().GitConf.RepoHttpUrl),
		BranchUrl:        fmt.Sprintf("%s/-/tree/%s", config.Get().GitConf.RepoHttpUrl, branchName),
		MergeRequestUrl:  fmt.Sprintf(""),
		SecretSealedYaml: string(secretSealedYaml),
	}, nil
}
