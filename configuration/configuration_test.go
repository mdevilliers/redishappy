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
