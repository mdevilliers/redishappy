package configuration

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Configuration struct {
	Clusters []Cluster
	HAProxy HAProxy
	Sentinels []Sentinel
}

func LoadFromFile(filePath string) (*Configuration, error) {
	configuration := new(Configuration)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &configuration)
	if err != nil {
		panic(err)
	}
	return configuration, nil
}

func (config *Configuration) String() (string) {
	e, _ := json.Marshal(config)
	return string(e[:])
}

func(config *Configuration) FindClusterByName (name string) (*Cluster,error) {

	for _, cluster := range config.Clusters {
    	if cluster.Name == name {
    		return &cluster, nil
    	}
	}
	return &Cluster{}, errors.New("Cluster not found")
}
