package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/util"
)

type HAProxyConfigContainsRequiredSections struct{}

func (c *HAProxyConfigContainsRequiredSections) Check(config configuration.Configuration) (bool, error) {

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

type CheckPermissionToWriteToHAProxyConfigFile struct{}

func (c *CheckPermissionToWriteToHAProxyConfigFile) Check(config configuration.Configuration) (bool, error) {

	file, err := os.OpenFile(config.HAProxy.OutputPath, os.O_RDWR|os.O_APPEND, 0660)

	defer file.Close()

	if err != nil {
		return false, fmt.Errorf("Configuration file at %s is not able to be opened.", config.HAProxy.OutputPath)
	}

	_, err = file.Write([]byte{' '})

	if err != nil {
		return false, fmt.Errorf("Configuration file at %s : error writing a empty byte.", config.HAProxy.OutputPath)
	}

	return true, nil
}

type CheckHAProxyTemplateFileExists struct{}

func (c *CheckHAProxyTemplateFileExists) Check(config configuration.Configuration) (bool, error) {

	exists := util.FileExists(config.HAProxy.TemplatePath)

	if exists {
		return true, nil
	}
	return false, fmt.Errorf("HAProxy Template file at %s : does not exist.", config.HAProxy.TemplatePath)
}
