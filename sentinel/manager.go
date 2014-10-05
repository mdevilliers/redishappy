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
	SentinelMarkedUp    = iota
	SentinelMarkedDown  = iota
	SentinelMarkedAlive = iota
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

func NewManager(switchmasterchannel chan types.MasterSwitchedEvent) *SentinelManager {
	events := make(chan SentinelEvent)
	requests := make(chan TopologyRequest)
	manager := &SentinelManager{eventsChannel: events,
		topologyRequestChannel: requests,
		switchmasterchannel:    switchmasterchannel,
		redisConnection:        redis.RadixRedisConnection{}}
	go loopEvents(events, requests, manager)
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

func (m *SentinelManager) Start(stateChannel chan types.MasterDetailsCollection, configuration *configuration.Configuration) {

	detailcollection := types.NewMasterDetailsCollection()

	for _, configuredSentinel := range configuration.Sentinels {

		client, err := m.newSentinelClient(configuredSentinel)

		if err != nil {
			logger.Info.Printf("Error starting sentinel (%s) client : %s", configuredSentinel.GetLocation(), err.Error())
			continue
		}

		go m.newMonitor(configuredSentinel)

		for _, clusterDetails := range configuration.Clusters {

			details, err := client.DiscoverMasterForCluster(clusterDetails.Name)

			if err != nil {
				continue
			}
			details.ExternalPort = clusterDetails.MasterPort
			// TODO : last one wins?
			detailcollection.AddOrReplace(details)

			// explore the cluster
			client.FindConnectedSentinels(clusterDetails.Name)

		}
	}

	stateChannel <- detailcollection

}

func (m *SentinelManager) newSentinelClient(sentinel types.Sentinel) (*SentinelClient, error) {

	m.Notify(&SentinelAdded{Sentinel: sentinel})
	client, err := NewSentinelClient(sentinel, m, m.redisConnection)

	if err != nil {
		logger.Error.Printf("Error starting sentinel client (%s) : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}
	return client, err
}

func (m *SentinelManager) newMonitor(sentinel types.Sentinel) (*Monitor, error) {

	monitor, err := NewMonitor(sentinel, m, m.redisConnection)

	if err != nil {
		logger.Error.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
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
		if _, ok := topologyState.Sentinels[uid]; !ok {

			info := &SentinelInfo{SentinelLocation: uid,
				LastUpdated:   time.Now().UTC(),
				KnownClusters: []string{},
				State:         SentinelMarkedUp}

			topologyState.Sentinels[uid] = info

			go m.newMonitor(sentinel)

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

		util.Schedule(func() { 
					_,err := m.newMonitor(sentinel) 
					if err != nil {
						m.Notify(&SentinelLost{Sentinel: sentinel})
					}

				}, time.Second * 5)
		
		logger.Trace.Printf("Sentinel lost : %s (scheduling new client and monitor).", util.String(topologyState))

	case *SentinelPing:
		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)
		currentInfo, ok := topologyState.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedAlive
			currentInfo.LastUpdated = time.Now().UTC()
			currentInfo.KnownClusters = e.Clusters
		}

	default:
		logger.Error.Println("Unknown sentinel event : ", util.String(e))
	}
}
