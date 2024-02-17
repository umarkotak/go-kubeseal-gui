package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		ControllerName      string             `json:"controller_name"`
		ControllerNamespace string             `json:"controller_namespace"`
		ClusterMap          map[string]Cluster `json:"cluster_map"`
	}

	Cluster struct {
		Alias             string   `json:"alias"`              // must unique
		Name              string   `json:"name"`               // the cluster name
		RegisteredSecrets []Secret `json:"registered_secrets"` // selected secrets that will be shown on ui
		AllSecrets        []Secret `json:"all_secrets"`        // set upon saving cluster, you need to re add if you want to sync the secrets
	}

	Secret struct {
		Name string `json:"name"`
	}
)

var (
	config = Config{
		ClusterMap: map[string]Cluster{},
	}
)

// Read from config file, then assign to config variable
func Load() error {
	file, err := os.Open("config.json")
	if err != nil {
		logrus.Error(err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}

	return nil
}

// Set config variable to file
func SetCluster(c Cluster) error {
	for _, oneCluster := range config.ClusterMap {
		if oneCluster.Name == c.Name {
			err := fmt.Errorf("cluster already exists")
			return err
		}
	}

	config.ClusterMap[c.Alias] = c

	err := saveToFile(config)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	Load()

	return nil
}

func SetClusterSecret(alias string, secrets []string) error {
	registeredSecrets := []Secret{}
	for _, oneSecret := range secrets {
		registeredSecrets = append(registeredSecrets, Secret{
			Name: oneSecret,
		})
	}

	cluster := config.ClusterMap[alias]

	cluster.RegisteredSecrets = registeredSecrets

	config.ClusterMap[alias] = cluster

	err := saveToFile(config)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	Load()

	return nil
}

func SetController(ctrlrName, ctrlrNamespace string) error {
	config.ControllerName = ctrlrName
	config.ControllerNamespace = ctrlrNamespace

	saveToFile(config)

	Load()

	return nil
}

func RemoveCluster(alias string) error {
	delete(config.ClusterMap, alias)

	err := saveToFile(config)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	Load()

	return nil
}

func saveToFile(c Config) error {
	file, err := os.Create("config.json")
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer file.Close()

	b, _ := json.Marshal(c)

	_, err = file.Write(b)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// Get config
func Get() Config {
	return config
}
