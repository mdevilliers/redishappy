package sentinel

import (
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
	NewSentinelClient(types.Sentinel) (*SentinelClient, error)
	NewMonitor(types.Sentinel) (*Monitor, error)
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

func (m *SentinelManager) NewSentinelClient(sentinel types.Sentinel) (*SentinelClient, error) {

	m.Notify(&SentinelAdded{Sentinel: sentinel})
	client, err := NewSentinelClient(sentinel, m, m.redisConnection)

	if err != nil {
		logger.Error.Printf("Error starting sentinel client (%s) : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}

	client.Start()
	return client, err
}

func (m *SentinelManager) NewMonitor(sentinel types.Sentinel) (*Monitor, error) {

	monitor, err := NewMonitor(sentinel, m.redisConnection)

	if err != nil {
		logger.Error.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}

	monitor.StartMonitoringMasterEvents(m.switchmasterchannel)
	return monitor, nil
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

func loopEvents(events chan SentinelEvent, topology chan TopologyRequest, m Manager) {
	for {
		select {
		case event := <-events:
			updateState(event, m)
		case read := <-topology:
			read.ReplyChannel <- topologyState
		}
	}
}

func updateState(event interface{}, m Manager) {

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

		util.Schedule(func() { m.NewSentinelClient(sentinel) }, time.Second*5)
		util.Schedule(func() { m.NewMonitor(sentinel) }, time.Second*5)
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
