package sentinel

import (
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"sync"
	"time"
)

const (
	SentinelMarkedUp    = 1
	SentinelMarkedDown  = 2
	SentinelMarkedAlive = 3
)

type Manager interface {
	Notify(event SentinelEvent)
	GetState(request TopologyRequest)
}

type SentinelManager struct {
	eventsChannel          chan SentinelEvent
	topologyRequestChannel chan TopologyRequest
	switchmasterchannel    chan types.MasterSwitchedEvent
	redisConnection        redis.RedisConnection
}

var topologyState = SentinelTopology{Sentinels: map[string]*SentinelInfo{}}
var statelock = &sync.Mutex{}

func NewManager(switchmasterchannel chan types.MasterSwitchedEvent, configuration *configuration.Configuration) *SentinelManager {

	events := make(chan SentinelEvent)
	requests := make(chan TopologyRequest)

	manager := &SentinelManager{eventsChannel: events,
		topologyRequestChannel: requests,
		switchmasterchannel:    switchmasterchannel,
		redisConnection:        redis.RadixRedisConnection{}}

	go loopEvents(events, requests, manager)

	for _, sentinel := range configuration.Sentinels {
		manager.Notify(&SentinelAdded{Sentinel: sentinel})
	}

	return manager
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.eventsChannel <- event
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.topologyRequestChannel <- request
}

func (m *SentinelManager) ClearState() {
	statelock.Lock()
	defer statelock.Unlock()
	topologyState = SentinelTopology{Sentinels: map[string]*SentinelInfo{}}
}

func (m *SentinelManager) GetTopology(stateChannel chan types.MasterDetailsCollection, configuration *configuration.Configuration) {

	topology := types.NewMasterDetailsCollection()

	for _, sentinel := range configuration.Sentinels {

		client, err := NewSentinelClient(sentinel, m.redisConnection)
		defer client.Close()
		
		if err != nil {
			logger.Info.Printf("Error starting sentinel (%s) client : %s", sentinel.GetLocation(), err.Error())
			continue
		}

		for _, clusterDetails := range configuration.Clusters {

			details, err := client.DiscoverMasterForCluster(clusterDetails.Name)

			if err != nil {
				continue
			}

			details.ExternalPort = clusterDetails.MasterPort
			// TODO : last one wins?
			topology.AddOrReplace(details)

			m.exploreSentinelTopology(client, clusterDetails.Name)
		}
	}
	stateChannel <- topology
}

func (m *SentinelManager) exploreSentinelTopology(client *SentinelClient, clustername string) {

	sentinels := client.FindConnectedSentinels(clustername)
	if len(sentinels) > 0 {
		m.notifySentinelsAreConnected(sentinels)
	}
}

func (m *SentinelManager) notifySentinelsAreConnected(sentinels []types.Sentinel) {
	for _, sentinel := range sentinels {
		m.Notify(&SentinelAdded{Sentinel: sentinel})
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

func loopEvents(events chan SentinelEvent, topology chan TopologyRequest, m *SentinelManager) {
	for {
		select {
		case event := <-events:
			updateState(event, m)
		case read := <-topology:
			read.ReplyChannel <- topologyState
		}
	}
}

func updateState(event interface{}, m *SentinelManager) {

	statelock.Lock()
	defer statelock.Unlock()

	switch e := event.(type) {
	case *SentinelAdded:

		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)

		//if we don't know about the sentinel start monitoring it
		if _, exists := topologyState.Sentinels[uid]; !exists {

			info := &SentinelInfo{SentinelLocation: uid,
				LastUpdated:   time.Now().UTC(),
				KnownClusters: []ClusterInfo{},
				State:         SentinelMarkedUp}

			topologyState.Sentinels[uid] = info

			go m.startNewMonitor(sentinel)

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

		util.Schedule(func() { m.startNewMonitor(sentinel) }, time.Second*5)

		logger.Trace.Printf("Sentinel lost : %s (scheduling new client and monitor).", util.String(topologyState))

	case *SentinelPing:
		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)
		currentInfo, ok := topologyState.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedAlive
			currentInfo.LastUpdated = time.Now().UTC()
			// currentInfo.KnownClusters = e.Clusters
		}

	default:
		logger.Error.Println("Unknown sentinel event : ", util.String(e))
	}
}
