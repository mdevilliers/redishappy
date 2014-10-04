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

	logger.InitLogging(logPath)

	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	sentinelManager := sentinel.NewManager(switchmasterchannel)

	go loopSentinelEvents(flipper, switchmasterchannel)
	go startMonitoring(flipper, sentinelManager, configuration)

	initApiServer(sentinelManager)
}

func initApiServer(manager *sentinel.SentinelManager) {

	logger.Info.Print("hosting json endpoint.")

	pongApi := api.PingApi{}
	sentinelApi := api.SentinelApi{Manager: manager}

	goji.Get("/api/ping", pongApi.Get)
	goji.Get("/api/sentinel", sentinelApi.Get)
	goji.Serve()
}

func startMonitoring(flipper types.FlipperClient, sentinelManager *sentinel.SentinelManager, configuration *configuration.Configuration) {

	detailsCollectionChannel := make(chan types.MasterDetailsCollection, 1)

	go sentinelManager.Start(detailsCollectionChannel, configuration)

	detailcollection := <-detailsCollectionChannel

	flipper.InitialiseRunningState(&detailcollection)
}

func loopSentinelEvents(flipper types.FlipperClient, switchmasterchannel chan types.MasterSwitchedEvent) {

	for switchEvent := range switchmasterchannel {
		flipper.Orchestrate(switchEvent)
	}
}
