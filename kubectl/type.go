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

	SecretSealed struct {
		APIVersion string               `yaml:"apiVersion"`
		Kind       string               `yaml:"kind"`
		Metadata   SecretSealedMetadata `yaml:"metadata"`
		Spec       SecretSealedSpec     `yaml:"spec"`
	}

	SecretSealedMetadata struct {
		CreationTimestamp interface{} `yaml:"creationTimestamp"`
		Name              string      `yaml:"name"`
		Namespace         string      `yaml:"namespace"`
	}

	SecretSealedSpec struct {
		EncryptedData map[string]string    `yaml:"encryptedData"`
		Template      SecretSealedTemplate `yaml:"template"`
	}

	SecretSealedTemplate struct {
		Metadata SecretSealedMetadata `yaml:"metadata"`
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

func (s *Secret) EncodeBase64() {
	for k, v := range s.Data {
		b64Decoded := base64.StdEncoding.EncodeToString([]byte(v))
		s.Data[k] = string(b64Decoded)
	}
}
