package configuration

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mdevilliers/redishappy/services/logger"
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

	foldInEnvironmentalVariables(&configuration)

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

func foldInEnvironmentalVariables(config *Configuration) {

	config.HAProxy.OutputPath = getEnvironmentalVariable("REDISHAPPY_HAPROXY_OUTPUT_PATH", config.HAProxy.OutputPath)
	config.HAProxy.TemplatePath = getEnvironmentalVariable("REDISHAPPY_HAPROXY_TEMPLATE_PATH", config.HAProxy.TemplatePath)
	config.HAProxy.ReloadCommand = getEnvironmentalVariable("REDISHAPPY_HAPROXY_RELOAD_CMD", config.HAProxy.ReloadCommand)
	overrideClusterConfiguration(config)
	overrideSentinelConfiguration(config)
}

func getEnvironmentalVariable(name string, defaultValue string) string {
	env := os.Getenv(name)
	if len(env) > 0 {
		logger.Info.Printf("Overriding with environmental variable %s=%s", name, env)
		return env
	}
	return defaultValue
}

func overrideClusterConfiguration(config *Configuration) {

	env := os.Getenv("REDISHAPPY_CLUSTERS")

	if len(env) > 0 {
		config.Clusters = []types.Cluster{}

		bits := strings.Split(env, ";")

		for _, clusterConfig := range bits {

			ok, host, port := fauxHostAndIpToBits(clusterConfig)

			if !ok {
				logger.Error.Panicf("Error parsing port REDISHAPPY_CLUSTERS : %s {%s}", env, clusterConfig)
			}
			config.Clusters = append(config.Clusters, types.Cluster{Name: host, ExternalPort: port})
		}
		logger.Info.Printf("Using environment override for cluster configuration REDISHAPPY_CLUSTERS : %s", env)
	}
}

func overrideSentinelConfiguration(config *Configuration) {

	env := os.Getenv("REDISHAPPY_SENTINELS")

	if len(env) > 0 {
		config.Sentinels = []types.Sentinel{}

		bits := strings.Split(env, ";")

		for _, sentinelConfig := range bits {

			ok, host, port := fauxHostAndIpToBits(sentinelConfig)

			if !ok {
				logger.Error.Panicf("Error parsing port REDISHAPPY_SENTINELS : %s {%s}", env, sentinelConfig)
			}
			config.Sentinels = append(config.Sentinels, types.Sentinel{Host: host, Port: port})

		}
		logger.Info.Printf("Using environment override for sentinel configuration REDISHAPPY_SENTINELS : %s", env)
	}
}

func fauxHostAndIpToBits(hostAndIp string) (bool, string, int) {
	bits := strings.Split(hostAndIp, ":")

	if len(bits) != 2 {
		return false, "", 0
	}

	port, err := strconv.Atoi(bits[1])

	if err != nil {
		return false, "", 0
	}

	return true, bits[0], port
}
