package types

import (
	"testing"
)

func TestCanFindNamedService(t *testing.T) {
	services := []Service{Service{Cluster: "one"}, Service{Cluster: "two"}}
	config := Consul{Services: services}

	found, err := config.FindByClusterName("two")

	if err != nil {
		t.Error("Should have found the custer called 'two'")
	}

	if found.Cluster != "two" {
		t.Error("Should have found the custer called 'two'")
	}
}

func TestErrorsIfCanNotFindNamedService(t *testing.T) {
	services := []Service{Service{Cluster: "one"}, Service{Cluster: "two"}}
	config := Consul{Services: services}

	_, err := config.FindByClusterName("three")

	if err == nil {
		t.Error("There isn't a cluster called 'three'")
	}

}
