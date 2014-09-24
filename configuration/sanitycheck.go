package configuration

import (
	"errors"
)

type SanityCheck interface {
	Check(config *Configuration) (bool, error)
}

type ConfigContainsAtLeastOneSentinelDefinition struct{}

func (c *ConfigContainsAtLeastOneSentinelDefinition) Check(config *Configuration) (bool, error) {

	if len(config.Sentinels) >= 1 {
		return true, nil
	}
	return false, errors.New("Configuration needs to contain at least one Sentinel.")
}

type ConfigContainsAtLeastOneClusterDefinition struct{}

func (c *ConfigContainsAtLeastOneClusterDefinition) Check(config *Configuration) (bool, error) {

	if len(config.Clusters) >= 1 {
		return true, nil
	}
	return false, errors.New("Configuration needs to contain at least one Cluster.")
}

type ConfigContainsRequiredSections struct{}

func (c *ConfigContainsRequiredSections) Check(config *Configuration) (bool, error) {

	if config.Clusters == nil {
		return false, errors.New("Configuration doesn't contain a 'Clusters' configuration.")
	}
	if config.Sentinels == nil {
		return false, errors.New("Configuration doesn't contain a 'Sentinels' configuration.")
	}

	if config.HAProxy.TemplatePath == "" {
		return false, errors.New("Configuration doesn't contain a 'HAProxy.TemplatePath' configuration.")
	}
	if config.HAProxy.OutputPath == "" {
		return false, errors.New("Configuration doesn't contain a 'HAProxy.OutputPath' configuration.")
	}
	if config.HAProxy.ReloadCommand == "" {
		return false, errors.New("Configuration doesn't contain a 'HAProxy.ReloadCommand' configuration.")
	}
	return true, nil
}
