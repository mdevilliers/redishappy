package main

import (
	"flag"
	"github.com/mdevilliers/redishappy"
	"github.com/mdevilliers/redishappy/configuration"
	"log"
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

	ok, errors := config.SanityCheckConfiguration(&configuration.ConfigContainsRequiredSections{}, &configuration.HAProxyConfigContainsRequiredSections{}, &configuration.CheckPermissionToWriteToHAProxyConfigFile{})

	if !ok {

		for _, errorAsStr := range errors {
			log.Print(errorAsStr)
		}

		log.Panic("Configuration fails checks")
	}

	flipper := NewHAProxyFlipper(config)
	redishappy.NewRedisHappyEngine(flipper, config, logPath)
}
