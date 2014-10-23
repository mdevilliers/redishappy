package main

import (
	"flag"
	"log"

	"github.com/mdevilliers/redishappy"
	"github.com/mdevilliers/redishappy/configuration"
)

var configFile string
var logPath string

func init() {
	flag.StringVar(&configFile, "config", "config.json", "Full path of the configuration JSON file.")
	flag.StringVar(&logPath, "log", "log", "Folder for the logging folder.")
}

func main() {

	flag.Parse()

	config, err := configuration.LoadFromFile(configFile)

	if err != nil {
		log.Panicf("Error opening config file : %s", err.Error())
	}

	sane, errors := config.GetCurrentConfiguration().SanityCheckConfiguration(&configuration.ConfigContainsRequiredSections{},
		&HAProxyConfigContainsRequiredSections{},
		&CheckPermissionToWriteToHAProxyConfigFile{},
		&CheckHAProxyTemplateFileExists{})

	if !sane {

		for _, errorAsStr := range errors {
			log.Print(errorAsStr)
		}

		log.Panic("Configuration fails checks")
	}

	flipper := NewHAProxyFlipper(config)
	redishappy.NewRedisHappyEngine(flipper, config, logPath)
}
