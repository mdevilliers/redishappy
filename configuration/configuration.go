package configuration

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Clusters []Cluster
	HAProxy HAProxy
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

func (config *Configuration) String() (string,error) {
	configurationStr, err := json.Marshal(config)
	if err != nil { 
		return "", err
	}
	return string(configurationStr[:]), nil
}
