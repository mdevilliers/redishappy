package redishappy

import (
	"github.com/mdevilliers/redishappy/api"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type Muxonate func(*web.Mux)

type RedisHappyEngine struct {
	cm *configuration.ConfigurationManager
	sm *sentinel.SentinelManager
}

func NewRedisHappyEngine(flipper types.FlipperClient, cm *configuration.ConfigurationManager) *RedisHappyEngine {
	logger.Info.Print("pre mux")
	masterEvents := make(chan types.MasterSwitchedEvent)
	sentinelManager := sentinel.NewManager(masterEvents, cm)

	go loopSentinelEvents(flipper, masterEvents)
	go intiliseTopology(flipper, sentinelManager)

	return &RedisHappyEngine{
		cm: cm,
		sm: sentinelManager,
	}
}

func (r *RedisHappyEngine) ConfigureHandlersAndServe(fn Muxonate) {
	initApiServer(r.sm, r.cm, fn)
}

func (r *RedisHappyEngine) Serve() {
	initApiServer(r.sm, r.cm, func(_ *web.Mux) {})
}

func loopSentinelEvents(flipper types.FlipperClient, masterEvents chan types.MasterSwitchedEvent) {

	for event := range masterEvents {
		flipper.Orchestrate(event)
	}
}

func intiliseTopology(flipper types.FlipperClient, sentinelManager *sentinel.SentinelManager) {

	topology := sentinelManager.GetCurrentTopology()
	flipper.InitialiseRunningState(&topology)
}

func initApiServer(manager *sentinel.SentinelManager, cm *configuration.ConfigurationManager, muxonator Muxonate) {

	logger.Info.Print("hosting json endpoint.")

	pongApi := api.PingApi{}
	sentinelApi := api.SentinelApi{Manager: manager}
	configurationApi := api.ConfigurationApi{ConfigurationManager: cm}
	topologyApi := api.TopologyApi{Manager: manager}

	goji.Get("/api/ping", pongApi.Get)
	goji.Get("/api/sentinels", sentinelApi.Get)
	goji.Get("/api/configuration", configurationApi.Get)
	goji.Get("/api/topology", topologyApi.Get)

	muxonator(goji.DefaultMux)

	goji.Serve()
}
