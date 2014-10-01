package main

import (
	"flag"
	"github.com/mdevilliers/redishappy/api"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/services/haproxy"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"github.com/zenazn/goji"
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

	logger.Info.Print("redis-happy started")

	configuration, err := configuration.LoadFromFile(configFile)

	if err != nil {
		logger.Error.Panicf("Error opening config file : %s", err.Error())
	}

	logger.Info.Printf("Parsed from config : %s\n", util.String(configuration))

	flipper := haproxy.NewFlipper(configuration)
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

	started := 0
	detailcollection := types.NewMasterDetailsCollection()

	for _, configuredSentinel := range configuration.Sentinels {

		client, err := sentinelManager.NewSentinelClient(configuredSentinel)

		if err != nil {
			logger.Info.Printf("Error starting sentinel (%s) client : %s", configuredSentinel.GetLocation(), err.Error())
			continue
		}

		sentinelManager.NewMonitor(configuredSentinel)
		started++

		for _, clusterDetails := range configuration.Clusters {

			details, err := client.DiscoverMasterForCluster(clusterDetails.Name)

			if err != nil {
				continue
			}
			details.ExternalPort = clusterDetails.MasterPort
			// TODO : last one wins?
			detailcollection.AddOrReplace(&details)

			// explore the cluster
			client.FindConnectedSentinels(clusterDetails.Name)
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
