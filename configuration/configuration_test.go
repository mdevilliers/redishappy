package configuration

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestParseValidConfiguration(t *testing.T) {
	config := GetTestConfigFile()

	configuration, err := parseConfiguration([]byte(config))

	if err != nil {
		t.Error("This is a valid configuration and shouldn't error : ", err.Error())
		return
	}

	if len(configuration.Clusters) != 2 {
		t.Error("There should be two clusters.")
		return
	}

	cluster, err := configuration.FindClusterByName("cluster one")

	if err != nil {
		t.Error("Couldn't find cluster one : ", err.Error())
		return
	}

	if cluster.Name != "cluster one" {
		t.Error("Wrong cluster found.")
		return
	}

	cluster, err = configuration.FindClusterByName("does-not-exist")
	if err == nil {
		t.Error("This should error - the cluster does not exist : ", err.Error())
		return
	}
}

func TestConfigurationManagerGivesCorrectConfig(t *testing.T) {
	config := GetTestConfigFile()

	configuration, _ := parseConfiguration([]byte(config))
	cm := NewConfigurationManager(configuration)
	parsedConfig := cm.GetCurrentConfiguration()

	RunTestConfigFileChecks(parsedConfig, t)
}

func TestParseInValidConfiguration(t *testing.T) {
	config := "{ xxx : 1 }"

	_, err := parseConfiguration([]byte(config))

	if err == nil {
		t.Error("This is an invalid configuration and should fail.")
	}
}

func TestNonExistentFile(t *testing.T) {

	_, err := LoadFromFile("does-not-exist.config")

	if err == nil {
		t.Error("File doesn't exist and no error thrown")
	}
}

func TestExistingFile(t *testing.T) {

	file, err := ioutil.TempFile("", "")
	defer file.Close()
	defer os.Remove(file.Name())

	if err != nil {
		t.Error("Error creating temp file")
	}

	config := GetTestConfigFile()

	file.Write([]byte(config))
	cm, err := LoadFromFile(file.Name())

	if err != nil {
		t.Error("Error reading configuration")
	}
	parsedConfig := cm.GetCurrentConfiguration()
	RunTestConfigFileChecks(parsedConfig, t)
}

func GetTestConfigFile() string {
	return `{
				  "Clusters" :[
				  {
				    "Name" : "cluster one",
				    "ExternalPort" : 6379
				  },
				  {
				    "Name" : "cluster two",
				    "ExternalPort" : 6380
				  }],
				  "Sentinels" : [ 
				      {"Host" : "192.168.0.20", "Port" : 26379},
				      {"Host" : "192.168.0.21", "Port" : 26379}
				  ]
			}`
}

func RunTestConfigFileChecks(parsedConfig Configuration, t *testing.T) {
	if len(parsedConfig.Clusters) != 2 {
		t.Error("There should be two clusters.")
		return
	}
	if len(parsedConfig.Sentinels) != 2 {
		t.Error("There should be two sentinels.")
		return
	}
}

func TestEnvironmentalVariablesSetInConfiguration(t *testing.T) {

	os.Setenv("REDISHAPPY_HAPROXY_RELOAD_CMD", "abc")

	config := GetTestConfigFile()
	configuration, _ := parseConfiguration([]byte(config))

	foldInEnvironmentalVariables(&configuration)

	if configuration.HAProxy.ReloadCommand != "abc" {
		t.Error("Environmental value - REDISHAPPY_HAPROXY_RELOAD_CMD not folded in.")
	}
}

func TestEnvironmentalVariablesSentinelsSetInConfiguration(t *testing.T) {

	os.Setenv("REDISHAPPY_SENTINELS", "172.17.42.1:26377;172.17.42.1:26378;172.17.42.1:26379")

	config := GetTestConfigFile()
	configuration, _ := parseConfiguration([]byte(config))

	foldInEnvironmentalVariables(&configuration)

	if len(configuration.Sentinels) != 3 {
		t.Error("Environmental value - REDISHAPPY_SENTINELS not folded in.")
	}
}

func TestEnvironmentalVariablesClustersSetInConfiguration(t *testing.T) {

	os.Setenv("REDISHAPPY_CLUSTERS", "abc:1234;def:1234;ghi:1234;klm:1234")

	config := GetTestConfigFile()
	configuration, _ := parseConfiguration([]byte(config))

	foldInEnvironmentalVariables(&configuration)

	if len(configuration.Clusters) != 4 {
		t.Error("Environmental value - REDISHAPPY_CLUSTERS not folded in.")
	}
}
