package redishappy

import (
	"github.com/mdevilliers/redishappy/api"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"github.com/zenazn/goji"
)

func NewRedisHappyEngine(flipper types.FlipperClient, configuration *configuration.Configuration, logPath string) {

	logger.InitLogging(logPath)
	logger.Info.Printf("Configuration : %s\n", util.String(configuration))

	masterEvents := make(chan types.MasterSwitchedEvent)
	sentinelManager := sentinel.NewManager(masterEvents, configuration)

	go loopSentinelEvents(flipper, masterEvents)
	go intiliseTopology(flipper, sentinelManager, configuration)

	initApiServer(sentinelManager)
}

func loopSentinelEvents(flipper types.FlipperClient, masterEvents chan types.MasterSwitchedEvent) {

	for event := range masterEvents {
		flipper.Orchestrate(event)
	}
}

func intiliseTopology(flipper types.FlipperClient, sentinelManager *sentinel.SentinelManager, configuration *configuration.Configuration) {

	stateChannel := make(chan types.MasterDetailsCollection)

	go sentinelManager.GetTopology(stateChannel, configuration)

	topology := <-stateChannel

	flipper.InitialiseRunningState(&topology)
}

func initApiServer(manager *sentinel.SentinelManager) {

	logger.Info.Print("hosting json endpoint.")

	pongApi := api.PingApi{}
	sentinelApi := api.SentinelApi{Manager: manager}

	goji.Get("/api/ping", pongApi.Get)
	goji.Get("/api/sentinel", sentinelApi.Get)

	goji.Serve()
}
