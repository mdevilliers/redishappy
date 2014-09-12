package configuration

import "testing"

func TestLoadinAndParseConfiguration(t *testing.T) {
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

	configuration, _ := ParseConfiguration([]byte(config))

	if len(configuration.Clusters) != 2 {
		t.Error("There should be two clusters")
	}

	cluster, err := configuration.FindClusterByName("cluster one")

	if err != nil {
		t.Error("Couldn't find cluster one")
	}

	if cluster.Name != "cluster one" {
		t.Error("Wrong cluster found")
	}
}

func TestNonExistentFile(t *testing.T) {
	
	_, err := LoadFromFile("does-not-exist.config") 

	if err == nil {
		t.Error("File doesn't exist and no error thrown")
	}
}