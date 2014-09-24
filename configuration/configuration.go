package configuration

import (
	"encoding/json"
	"errors"
	"github.com/mdevilliers/redishappy/types"
	"io/ioutil"
)

type Configuration struct {
	Clusters  []types.Cluster
	HAProxy   types.HAProxy
	Sentinels []types.Sentinel
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
		return nil, err
	}

	ok, err := SenseCheckConfiguration(configuration)

	if !ok {
		return nil, err
	}

	return configuration, err
}

func SenseCheckConfiguration(config *Configuration) (bool, error) {

	tests := []SanityCheck{&ConfigContainsRequiredSections{},
		&ConfigContainsAtLeastOneSentinelDefinition{},
		&ConfigContainsAtLeastOneClusterDefinition{}}

	for _, test := range tests {

		ok, err := test.Check(config)

		if !ok {
			return false, err
		}
	}

	return true, nil
}

func (config *Configuration) FindClusterByName(name string) (*types.Cluster, error) {

	for _, cluster := range config.Clusters {
		if cluster.Name == name {
			return &cluster, nil
		}
	}
	return &types.Cluster{}, errors.New("Cluster not found")
}
