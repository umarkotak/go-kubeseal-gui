package kubectl

import (
	"context"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
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

	return k8sNames, nil
}
