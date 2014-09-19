package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
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

type SentinelManager struct {
	eventsChannel          chan SentinelEvent
	topologyRequestChannel chan TopologyRequest
	switchmasterchannel    chan MasterSwitchedEvent
}

var topologyState = SentinelTopology{Sentinels: map[string]*SentinelInfo{}}
var statelock = &sync.Mutex{}

func NewManager(switchmasterchannel chan MasterSwitchedEvent) *SentinelManager {
	events := make(chan SentinelEvent)
	requests := make(chan TopologyRequest)
	go loopEvents(events, requests)
	return &SentinelManager{eventsChannel: events, topologyRequestChannel: requests, switchmasterchannel : switchmasterchannel}
}

func (m *SentinelManager) NewSentinelMonitor(sentinel types.Sentinel) (*SentinelHealthCheckerClient, error) {

	client, err := NewHealthCheckerClient(sentinel, m)

	if err != nil {
		logger.Info.Printf("Error starting health checker (%s) : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}

	m.Notify(&SentinelAdded{Sentinel: &sentinel})
	client.Start()

	pubsubclient, err := NewPubSubClient(sentinel)

	if err != nil {
		logger.Info.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}

	pubsubclient.StartMonitoringMasterEvents(m.switchmasterchannel)

	return client, err
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

func loopEvents(events chan SentinelEvent, topology chan TopologyRequest) {
	for {
		select {
		case event := <-events:
			updateState(event)
		case read := <-topology:
			read.ReplyChannel <- topologyState
		}
	}
}

func updateState(event interface{}) {

	statelock.Lock()
	defer statelock.Unlock()

	switch e := event.(type) {
	case *SentinelAdded:

		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)
		info := &SentinelInfo{SentinelLocation: uid,
			LastUpdated:   time.Now().UTC(),
			KnownClusters: []string{},
			State:         SentinelMarkedUp}

		topologyState.Sentinels[uid] = info
		logger.Trace.Printf("Sentinel added : %s", util.String(topologyState))
		
	case *SentinelLost:

		sentinel := e.GetSentinel()
		uid := topologyState.createKey(sentinel)
		currentInfo, ok := topologyState.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedDown
			currentInfo.LastUpdated = time.Now().UTC()
		}
		logger.Trace.Printf("Sentinel lost : %s", util.String(topologyState))

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
