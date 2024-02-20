package kubectl

import (
	"encoding/base64"
)

type (
	Secret struct {
		APIVersion string            `yaml:"apiVersion"`
		Kind       string            `yaml:"kind"`
		Metadata   Metadata          `yaml:"metadata"`
		Data       map[string]string `yaml:"data"`
	}

	Metadata struct {
		Name string `yaml:"name"`
	}
)

func (s *Secret) DecodeBase64() error {
	for k, v := range s.Data {
		b64Decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		s.Data[k] = string(b64Decoded)
	}
	return nil
}
