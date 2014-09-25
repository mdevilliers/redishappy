package main

import (
	"github.com/mdevilliers/redishappy/api"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/services/haproxy"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"github.com/zenazn/goji"
)

func main() {

	//TODO : configure from command line
	logPath := "log" //var/log/redis-happy")
	configFile := "config.json"

	logger.InitLogging(logPath)

	logger.Info.Print("redis-happy started")

	configuration, err := configuration.LoadFromFile(configFile)

	if err != nil {
		logger.Error.Panic(err)
	}

	logger.Info.Printf("Parsed from config : %s\n", util.String(configuration))

	flipper := haproxy.NewFlipper(configuration)

	switchmasterchannel := make(chan types.MasterSwitchedEvent)

	go loopSentinelEvents(flipper, switchmasterchannel)

	sentinelManager := sentinel.NewManager(switchmasterchannel)

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

	started := 0
	detailcollection := types.NewMasterDetailsCollection()

	for _, configuredSentinel := range configuration.Sentinels {

		client, err := sentinelManager.NewSentinelMonitor(configuredSentinel)

		if err != nil {
			logger.Info.Printf("Error starting sentinel (%s) healthchecker : %s", configuredSentinel.GetLocation(), err.Error())
		} else {
			started++

			for _, clusterDetails := range configuration.Clusters {

				details, err := client.DiscoverMasterForCluster(clusterDetails.Name)

				if err == nil {
					details.ExternalPort = clusterDetails.MasterPort
					// TODO : last one wins?
					detailcollection.AddOrReplace(&details)
				}
			}
		}
	}

	flipper.InitialiseRunningState(detailcollection)

	if started == 0 {
		logger.Info.Printf("WARNING : no sentinels are currently being monitored.")
	}
}

func loopSentinelEvents(flipper types.FlipperClient, switchmasterchannel chan types.MasterSwitchedEvent) {

	for switchEvent := range switchmasterchannel {
		flipper.Orchestrate(switchEvent)
	}
}
