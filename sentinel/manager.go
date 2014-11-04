package sentinel

import (
	"time"

	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

const (
	SentinelMarkedUp    = 1
	SentinelMarkedDown  = 2
	SentinelMarkedAlive = 3
)

const (
	SentinelReconnectionPeriod        = time.Second * 5
	SentinelTopologyExplorationPeriod = time.Second * 60
)

type Manager interface {
	Notify(event SentinelEvent)
}

type SentinelManager struct {
	switchmasterchannel  chan types.MasterSwitchedEvent
	redisConnection      redis.RedisConnection
	configurationManager *configuration.ConfigurationManager
	throttle             *Throttle
	state                SentinelState
}

func NewManager(switchmasterchannel chan types.MasterSwitchedEvent, cm *configuration.ConfigurationManager) *SentinelManager {

	unthrottled := make(chan types.MasterSwitchedEvent)
	throttle := NewThrottle(unthrottled, switchmasterchannel)

	manager := &SentinelManager{
		switchmasterchannel:  unthrottled,
		redisConnection:      redis.RadixRedisConnection{},
		configurationManager: cm,
		throttle:             throttle,
	}

	startMonitoringCallback := func(sentinel types.Sentinel) {
		go manager.exploreSentinel(sentinel)
		go manager.startNewMonitor(sentinel)
	}

	manager.state = NewSentinelState(startMonitoringCallback)

	go manager.bootstrap()
	return manager
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.state.Notify(event)
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.state.GetState(request)
}

func (m *SentinelManager) GetCurrentTopology() types.MasterDetailsCollection {
	stateChannel := make(chan types.MasterDetailsCollection)
	go m.getTopology(stateChannel)
	return <-stateChannel
}

func (m *SentinelManager) exploreSentinel(sentinel types.Sentinel) {

	client, err := NewSentinelClient(sentinel, m.redisConnection)

	if err != nil {

		logger.Error.Printf("Error starting sentinel (%s) client : %s", sentinel.GetLocation(), err.Error())
		m.Notify(&SentinelLost{Sentinel: sentinel})

		return
	}
	defer client.Close()

	knownClusters := client.FindKnownClusters()

	m.Notify(&SentinelClustersMonitoredUpdate{Sentinel: sentinel, Clusters: knownClusters})

	for _, clustername := range knownClusters {

		sentinels := client.FindConnectedSentinels(clustername)

		for _, connectedsentinel := range sentinels {
			m.Notify(&SentinelAdded{Sentinel: connectedsentinel})
		}
	}
}

func (m *SentinelManager) startNewMonitor(sentinel types.Sentinel) {

	monitor, err := NewMonitor(sentinel, m, m.redisConnection)

	if err != nil {
		logger.Error.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
		m.Notify(&SentinelLost{Sentinel: sentinel})
		return
	}

	go monitor.StartMonitoringMasterEvents(m.switchmasterchannel)
}

func (m *SentinelManager) getTopology(stateChannel chan types.MasterDetailsCollection) {

	topology := types.NewMasterDetailsCollection()
	configuration := m.configurationManager.GetCurrentConfiguration()

	for _, sentinel := range configuration.Sentinels {
		client, err := NewSentinelClient(sentinel, m.redisConnection)

		if err != nil {
			logger.Info.Printf("Error starting sentinel (%s) client : %s", sentinel.GetLocation(), err.Error())
			continue
		}
		defer client.Close()

		for _, clusterDetails := range configuration.Clusters {

			details, err := client.DiscoverMasterForCluster(clusterDetails.Name)
			if err != nil {
				continue
			}

			details.ExternalPort = clusterDetails.ExternalPort
			// TODO : last one wins?
			topology.AddOrReplace(details)
		}
	}
	stateChannel <- topology
}

func (m *SentinelManager) bootstrap() {

	configuration := m.configurationManager.GetCurrentConfiguration()

	for _, sentinel := range configuration.Sentinels {
		m.Notify(&SentinelAdded{Sentinel: sentinel})
	}

	util.Schedule(func() { m.bootstrap() }, SentinelTopologyExplorationPeriod)
}
