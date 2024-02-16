package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		ClusterMap map[string]Cluster `json:"cluster_map"`
	}

	Cluster struct {
		Alias             string   `json:"alias"` // must unique
		Name              string   `json:"name"`  // the cluster name
		RegisteredSecrets []Secret `json:"registered_secrets"`
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
	config.ClusterMap[c.Alias] = c

	file, err := os.Create("config.json")
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer file.Close()

	b, _ := json.Marshal(config)

	_, err = file.Write(b)
	if err != nil {
		logrus.Error(err)
		return err
	}

	Load()

	return nil
}

// Get config
func Get() Config {
	return config
}
