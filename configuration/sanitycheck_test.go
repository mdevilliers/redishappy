package configuration

import (
	"testing"

	"github.com/mdevilliers/redishappy/types"
)

func TestObviousMisConfiguration(t *testing.T) {

	clusters := []types.Cluster{types.Cluster{Name: "", ExternalPort: 1234}} // no name
	sentinels := []types.Sentinel{types.Sentinel{Host: "192.168.0.20", Port: 26379}}

	config := &Configuration{Clusters: clusters, Sentinels: sentinels}

	sane, _ := config.SanityCheckConfiguration(&CheckForObviousMisConfiguration{})

	if sane {
		t.Error("Cluster has no name - not sane.")
	}

	config.Clusters[0].Name = "one"
	config.Clusters[0].ExternalPort = 0

	sane, _ = config.SanityCheckConfiguration(&CheckForObviousMisConfiguration{})

	if sane {
		t.Error("Cluster has no external port - not sane.")
	}

	config.Clusters[0].ExternalPort = 1234
	config.Sentinels[0].Host = ""

	sane, _ = config.SanityCheckConfiguration(&CheckForObviousMisConfiguration{})

	if sane {
		t.Error("Sentinel has no host address- not sane.")
	}

	config.Sentinels[0].Host = "192.168.0.20"
	config.Sentinels[0].Port = 0

	sane, _ = config.SanityCheckConfiguration(&CheckForObviousMisConfiguration{})

	if sane {
		t.Error("Sentinel has no port - not sane.")
	}

	config.Sentinels[0].Port = 1234

	sane, _ = config.SanityCheckConfiguration(&CheckForObviousMisConfiguration{})

	if !sane {
		t.Error("Configuration is sane.")
	}
}

func TestSanityCheckBasicUsage(t *testing.T) {

	clusters := []types.Cluster{types.Cluster{Name: "one", ExternalPort: 1234}}
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
