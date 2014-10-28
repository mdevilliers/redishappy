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
	eventsChannel          chan SentinelEvent
	topologyRequestChannel chan TopologyRequest
	switchmasterchannel    chan types.MasterSwitchedEvent
	redisConnection        redis.RedisConnection
	configurationManager   *configuration.ConfigurationManager
	throttle               *Throttle
}

var topologyState = SentinelTopology{Sentinels: map[string]*SentinelInfo{}}

func NewManager(switchmasterchannel chan types.MasterSwitchedEvent, cm *configuration.ConfigurationManager) *SentinelManager {

	events := make(chan SentinelEvent)
	requests := make(chan TopologyRequest)
	unthrottled := make(chan types.MasterSwitchedEvent)

	throttle := NewThrottle(unthrottled, switchmasterchannel)

	manager := &SentinelManager{eventsChannel: events,
		topologyRequestChannel: requests,
		switchmasterchannel:    unthrottled,
		redisConnection:        redis.RadixRedisConnection{},
		configurationManager:   cm,
		throttle:               throttle}

	go manager.loopEvents(events, requests)
	go manager.bootstrap()
	return manager
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.eventsChannel <- event
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.topologyRequestChannel <- request
}

func (m *SentinelManager) ClearState() {
	topologyState = SentinelTopology{Sentinels: map[string]*SentinelInfo{}}
}

func (m *SentinelManager) GetCurrentTopology() types.MasterDetailsCollection {
	stateChannel := make(chan types.MasterDetailsCollection)
	go m.getTopology(stateChannel)
	return <-stateChannel
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

func (m *SentinelManager) exploreSentinelTopology(sentinel types.Sentinel) {

	client, err := NewSentinelClient(sentinel, m.redisConnection)

	if err != nil {
		logger.Info.Printf("Error starting sentinel (%s) client : %s", sentinel.GetLocation(), err.Error())
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

func (m *SentinelManager) startNewMonitor(sentinel types.Sentinel) (*Monitor, error) {

	monitor, err := NewMonitor(sentinel, m, m.redisConnection)

	if err != nil {
		logger.Error.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
		m.Notify(&SentinelLost{Sentinel: sentinel})
		return nil, err
	}

	go monitor.StartMonitoringMasterEvents(m.switchmasterchannel)

	return monitor, nil
}

func (m *SentinelManager) loopEvents(events chan SentinelEvent, topology chan TopologyRequest) {
	for {
		select {
		case event := <-events:
			m.updateState(event)
		case read := <-topology:
			read.ReplyChannel <- topologyState
		}
	}
}

func (m *SentinelManager) updateState(event interface{}) {

	switch e := event.(type) {
	case *SentinelAdded:

		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)

		//if we don't know about the sentinel start monitoring it
		if _, exists := topologyState.Sentinels[uid]; !exists {

			info := &SentinelInfo{SentinelLocation: uid,
				LastUpdated: time.Now().UTC(),
				State:       SentinelMarkedUp}

			topologyState.Sentinels[uid] = info

			go m.startNewMonitor(sentinel)
			go m.exploreSentinelTopology(sentinel)

			logger.Trace.Printf("Sentinel added : %s", util.String(topologyState))
		}

	case *SentinelLost:

		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)
		currentInfo, ok := topologyState.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedDown
			currentInfo.LastUpdated = time.Now().UTC()
		}

		util.Schedule(func() { m.startNewMonitor(sentinel) }, SentinelReconnectionPeriod)

		logger.Trace.Printf("Sentinel lost : %s (scheduling new client and monitor).", util.String(topologyState))

	case *SentinelPing:
		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)
		currentInfo, ok := topologyState.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedAlive
			currentInfo.LastUpdated = time.Now().UTC()
		}

	case *SentinelClustersMonitoredUpdate:
		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)

		if info, exists := topologyState.Sentinels[uid]; exists {

			info.Clusters = e.Clusters
		}

	default:
		logger.Error.Println("Unknown sentinel event : ", util.String(e))
	}
}
