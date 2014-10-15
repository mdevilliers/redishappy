package redishappy

import (
	"github.com/mdevilliers/redishappy/api"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/zenazn/goji"
)

func NewRedisHappyEngine(flipper types.FlipperClient, cm *configuration.ConfigurationManager, logPath string) {

	logger.InitLogging(logPath)
	config := cm.GetCurrentConfiguration()

	masterEvents := make(chan types.MasterSwitchedEvent)
	sentinelManager := sentinel.NewManager(masterEvents, config)

	go loopSentinelEvents(flipper, masterEvents)
	go intiliseTopology(flipper, sentinelManager, config)

	initApiServer(sentinelManager, cm)
}

func loopSentinelEvents(flipper types.FlipperClient, masterEvents chan types.MasterSwitchedEvent) {

	for event := range masterEvents {
		flipper.Orchestrate(event)
	}
}

func intiliseTopology(flipper types.FlipperClient, sentinelManager *sentinel.SentinelManager, configuration configuration.Configuration) {

	stateChannel := make(chan types.MasterDetailsCollection)

	go sentinelManager.GetTopology(stateChannel, configuration)

	topology := <-stateChannel

	flipper.InitialiseRunningState(&topology)
}

func initApiServer(manager *sentinel.SentinelManager, cm *configuration.ConfigurationManager) {

	logger.Info.Print("hosting json endpoint.")

	pongApi := api.PingApi{}
	sentinelApi := api.SentinelApi{Manager: manager}
	configurationApi := api.ConfigurationApi{ConfigurationManager: cm}

	goji.Get("/api/ping", pongApi.Get)
	goji.Get("/api/sentinel", sentinelApi.Get)
	goji.Get("/api/configuration", configurationApi.Get)

	goji.Serve()
}
