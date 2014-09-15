package configuration

import "testing"

func TestParseValidConfiguration(t *testing.T) {
	config := `{
				  "Clusters" :[
				  {
				    "Name" : "cluster one",
				    "MasterPort" : 6379,
				    "SlavePorts" : [16379, 16380]  
				  },
				  {
				    "Name" : "cluster two",
				    "MasterPort" : 6380
				  }]
			}`

	configuration, err := ParseConfiguration([]byte(config))

	if err != nil {
		t.Error("This is a valid configuration and shouldn't error")
	}

	if len(configuration.Clusters) != 2 {
		t.Error("There should be two clusters.")
	}

	cluster, err := configuration.FindClusterByName("cluster one")

	if err != nil {
		t.Error("Couldn't find cluster one.")
	}

	if cluster.Name != "cluster one" {
		t.Error("Wrong cluster found.")
	}

	cluster, err = configuration.FindClusterByName("does-not-exist")
	if err == nil {
		t.Error("This should error - the cluster does not exist.")
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