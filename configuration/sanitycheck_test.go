package configuration

import (
	"testing"

	"github.com/mdevilliers/redishappy/types"
)

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
