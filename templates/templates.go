package templates

import (
	"fmt"
	"text/template"

	"github.com/sirupsen/logrus"
)

func Get(fileNames ...string) (*template.Template, error) {
	fileFullName := []string{}
	for _, fileName := range fileNames {
		fileFullName = append(fileFullName, fmt.Sprintf("templates/%s", fileName))
	}

	pt, err := template.New("").ParseFiles(fileFullName...)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return pt, nil
}
