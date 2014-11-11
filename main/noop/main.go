package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/mdevilliers/redishappy"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/zenazn/goji/web"
)

var configFile string
var logPath string

func init() {
	flag.StringVar(&configFile, "config", "config.json", "Full path of the configuration JSON file.")
	flag.StringVar(&logPath, "log", "log", "Folder for the logging folder.")
}

func main() {

	flag.Parse()
	logger.InitLogging(logPath)

	config, err := configuration.LoadFromFile(configFile)

	if err != nil {
		logger.Error.Panicf("Error opening config file : %s", err.Error())
	}

	sane, errors := config.GetCurrentConfiguration().SanityCheckConfiguration(&configuration.ConfigContainsRequiredSections{})

	if !sane {

		for _, errorAsStr := range errors {
			logger.Error.Print(errorAsStr)
		}

		logger.Error.Panic("Configuration fails checks")
	}

	flipper := NewNoOpFlipper()
	engine := redishappy.NewRedisHappyEngine(flipper, config)
	engine.ConfigureHandlersAndServe(AddHandlers)
}

// example handler
func AddHandlers(mux *web.Mux) {
	logger.Info.Print("muxed!")
	mux.Get("/api/xxxx", hello)
}
func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello!")
}
