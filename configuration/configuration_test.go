package configuration

import (
	"github.com/mdevilliers/redishappy/types"
	"testing"
)

func TestParseValidConfiguration(t *testing.T) {
	config := `{
				  "Clusters" :[
				  {
				    "Name" : "cluster one",
				    "MasterPort" : 6379
				  },
				  {
				    "Name" : "cluster two",
				    "MasterPort" : 6380
				  }],
				  "Sentinels" : [ 
				      {"Host" : "192.168.0.20", "Port" : 26379},
				      {"Host" : "192.168.0.21", "Port" : 26379}
				  ]
			}`

	configuration, err := ParseConfiguration([]byte(config))

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

func TestParseInValidConfiguration(t *testing.T) {
	config := "{ xxx : 1 }"

	_, err := ParseConfiguration([]byte(config))

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

func TestSanityCheckBasicUsage(t *testing.T) {

	clusters := []types.Cluster{types.Cluster{Name: "one", MasterPort: 1234}}
	sentinels := []types.Sentinel{types.Sentinel{Host: "192.168.0.20", Port: 26379}}

	config := &Configuration{Clusters: clusters, Sentinels: sentinels}

	sane, errors := config.SanityCheckConfiguration(&ConfigContainsRequiredSections{})

	if !sane {
		t.Errorf("This is a valid sanity checked configuration : %t, %d", sane, len(errors))
	}

	config.Sentinels = []types.Sentinel{}

	sane, errors = config.SanityCheckConfiguration(&ConfigContainsRequiredSections{})

	if sane {
		t.Errorf("Configuration has no sentinels configured : %t, %d", sane, len(errors))
	}

	config.Sentinels = nil

	sane, errors = config.SanityCheckConfiguration(&ConfigContainsRequiredSections{})

	if sane {
		t.Errorf("Configuration has no sentinels configured : %t, %d", sane, len(errors))
	}

	config.Sentinels = sentinels
	config.Clusters = []types.Cluster{}

	sane, errors = config.SanityCheckConfiguration(&ConfigContainsRequiredSections{})

	if sane {
		t.Errorf("Configuration has no clusters configured : %t, %d", sane, len(errors))
	}

	config.Clusters = nil

	sane, errors = config.SanityCheckConfiguration(&ConfigContainsRequiredSections{})

	if sane {
		t.Errorf("Configuration has no clusters configured : %t, %d", sane, len(errors))
	}
}
