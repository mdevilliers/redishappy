package configuration

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Configuration struct {
	Clusters  []Cluster
	HAProxy   HAProxy
	Sentinels []Sentinel
}

func LoadFromFile(filePath string) (*Configuration, error) {

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return &Configuration{}, err
	}
	return ParseConfiguration(content)
}

func ParseConfiguration(configurationAsJson []byte) (*Configuration, error) {
	configuration := new(Configuration)
	err := json.Unmarshal(configurationAsJson, &configuration)
	if err != nil {
		panic(err)
	}
	return configuration, nil
}

func (config *Configuration) FindClusterByName(name string) (*Cluster, error) {

	for _, cluster := range config.Clusters {
		if cluster.Name == name {
			return &cluster, nil
		}
	}
	return &Cluster{}, errors.New("Cluster not found")
}
