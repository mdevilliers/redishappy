package configuration

import (
	"errors"
	"fmt"
)

type SanityCheck interface {
	Check(config Configuration) (bool, error)
}

type ConfigContainsRequiredSections struct{}

func (c *ConfigContainsRequiredSections) Check(config Configuration) (bool, error) {

	if config.Clusters == nil {
		return false, errors.New("Configuration doesn't contain a 'Clusters' configuration.")
	}
	if len(config.Clusters) == 0 {
		return false, errors.New("Configuration needs to contain at least one Cluster.")
	}
	if config.Sentinels == nil {
		return false, errors.New("Configuration doesn't contain a 'Sentinels' configuration.")
	}
	if len(config.Sentinels) == 0 {
		return false, errors.New("Configuration needs to contain at least one Sentinel.")
	}

	return true, nil
}

type CheckForObviousMisConfiguration struct{}

func (c *CheckForObviousMisConfiguration) Check(config Configuration) (bool, error) {

	for cluster := range config.Clusters {

		if cluster.ExternalPort == 0 {
			return false, errors.New(fmt.Sprintf("Cluster %s configured with port 0", cluster.Name))
		}

		if cluster.Name == "" {
			return false, errors.New("Cluster configured without name")
		}
	}

	for sentinel := range config.Sentinels {

		if sentinel.Port == 0 {
			return false, errors.New(fmt.Sprintf("Sentinel %s configured with port 0", sentinel.Host))
		}

		if sentinel.Host == "" {
			return false, errors.New("Sentinel configured without host address")
		}
	}

	return true, nil
}