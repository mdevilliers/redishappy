package configuration

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/mdevilliers/redishappy/types"
)

type ConfigurationManager struct {
	config             Configuration
	getChannel         chan GetConfigCommand
	pathToOriginalFile string
}

type Configuration struct {
	Clusters  []types.Cluster  `json:"clusters"`
	Sentinels []types.Sentinel `json:"sentinels"`
	Consul    types.Consul     `json:"consul,omitempty"`
	HAProxy   types.HAProxy    `json:"HAProxy,omitempty"`
}

type GetConfigCommand struct {
	returnChannel chan Configuration
}

func (c *ConfigurationManager) GetCurrentConfiguration() Configuration {

	returnChannel := make(chan Configuration)
	c.getChannel <- GetConfigCommand{returnChannel: returnChannel}

	return <-returnChannel
}

func NewConfigurationManager(config Configuration) *ConfigurationManager {

	get := make(chan GetConfigCommand)
	cm := &ConfigurationManager{config: config, getChannel: get}

	go cm.loop(get)
	return cm
}

func (cm *ConfigurationManager) loop(get chan GetConfigCommand) {
	for {
		select {
		case getMessage := <-get:
			getMessage.returnChannel <- cm.config
		}
	}
}

func LoadFromFile(filePath string) (*ConfigurationManager, error) {

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	configuration, err := parseConfiguration(content)
	if err != nil {
		return nil, err
	}

	cm := NewConfigurationManager(configuration)
	cm.pathToOriginalFile = filePath

	return cm, nil
}

func parseConfiguration(configurationAsJson []byte) (Configuration, error) {

	configuration := Configuration{}
	err := json.Unmarshal(configurationAsJson, &configuration)
	if err != nil {
		return configuration, err
	}

	return configuration, nil
}

func (c Configuration) SanityCheckConfiguration(tests ...SanityCheck) (bool, []string) {

	errorlist := []string{}
	isSane := true

	for _, test := range tests {

		ok, err := test.Check(c)
		if !ok {
			errorlist = append(errorlist, err.Error())
			isSane = false
		}
	}

	return isSane, errorlist
}

func (config Configuration) FindClusterByName(name string) (*types.Cluster, error) {

	for _, cluster := range config.Clusters {
		if cluster.Name == name {
			return &cluster, nil
		}
	}
	return &types.Cluster{}, errors.New("Cluster not found")
}
