package configuration

import (
	"errors"
)

type SanityCheck interface {
	Check(config *Configuration) (bool, error)
}

type ConfigContainsRequiredSections struct{}

func (c *ConfigContainsRequiredSections) Check(config *Configuration) (bool, error) {

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
