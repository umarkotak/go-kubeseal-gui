package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/config"
	"gopkg.in/yaml.v2"
)

func GetContexts(ctx context.Context) ([]string, error) {
	k8sNames := []string{}

	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return k8sNames, err
	}

	k8sNames = strings.Split(string(output), "\n")

	filteredK8sNames := []string{}

	for _, k8sName := range k8sNames {
		if k8sName == "" {
			continue
		}

		filteredK8sNames = append(filteredK8sNames, k8sName)
	}

	return filteredK8sNames, nil
}

func GetSecretsName(ctx context.Context) ([]config.Secret, error) {
	secretNames := []string{}

	cmd := exec.Command("kubectl", "get", "secrets", "-o", "name")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("%s - %s", err.Error(), stderr.String())
		logrus.WithContext(ctx).Error(err)
		return []config.Secret{}, err
	}

	secretNames = strings.Split(string(output), "\n")

	filteredSecretNames := []config.Secret{}

	for _, name := range secretNames {
		if name == "" {
			continue
		}

		filteredSecretNames = append(filteredSecretNames, config.Secret{
			Name: strings.ReplaceAll(name, "secret/", ""),
		})
	}

	return filteredSecretNames, nil
}

func UseContext(ctx context.Context, name string) error {
	cmd := exec.Command("kubectl", "config", "use-context", name)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	_, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("%s - %s", err.Error(), stderr.String())
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"name": name,
		}).Error(err)
		return err
	}

	return nil
}

func GetSecretYaml(ctx context.Context, secretName string) (Secret, error) {
	kubeCmd := exec.Command("kubectl", "get", "secret", secretName, "-o", "yaml")

	var stderr bytes.Buffer
	kubeCmd.Stderr = &stderr

	output, err := kubeCmd.Output()
	if err != nil {
		err := fmt.Errorf(fmt.Sprint(err) + ": " + stderr.String())
		return Secret{}, err
	}

	secret := Secret{}

	err = yaml.Unmarshal(output, &secret)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return Secret{}, err
	}

	return secret, nil
}
