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
	NewSentinelMonitor(types.Sentinel) (*SentinelHealthCheckerClient, error)
	ScheduleNewHealthChecker(sentinel types.Sentinel)
}

type SentinelManager struct {
	eventsChannel          chan SentinelEvent
	topologyRequestChannel chan TopologyRequest
	switchmasterchannel    chan types.MasterSwitchedEvent
}

var topologyState = SentinelTopology{Sentinels: map[string]*SentinelInfo{}}
var statelock = &sync.Mutex{}

func NewManager(switchmasterchannel chan types.MasterSwitchedEvent) *SentinelManager {
	events := make(chan SentinelEvent)
	requests := make(chan TopologyRequest)
	manager := &SentinelManager{eventsChannel: events, topologyRequestChannel: requests, switchmasterchannel: switchmasterchannel}
	go loopEvents(events, requests, manager)
	return manager
}

func (m *SentinelManager) NewSentinelMonitor(sentinel types.Sentinel) (*SentinelHealthCheckerClient, error) {

	m.Notify(&SentinelAdded{Sentinel: sentinel})
	redisConnection := &redis.RadixRedisConnection{}
	client, err := NewHealthCheckerClient(sentinel, m, redisConnection)

	if err != nil {
		logger.Info.Printf("Error starting health checker (%s) : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}

	client.Start()

	pubsubclient, err := NewPubSubClient(sentinel, &redis.RadixRedisConnection{})

	if err != nil {
		logger.Info.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
		return nil, err
	}

	pubsubclient.StartMonitoringMasterEvents(m.switchmasterchannel)

	return client, err
}

func (m *SentinelManager) DiscoverMasterForCluster(clusterName string) types.MasterDetails {

	// dummy implementation as a strawman
	if clusterName == "secure" {
		return types.MasterDetails{Name: clusterName, Ip: "1.1.1.1", Port: 1234}
	}
	return types.MasterDetails{Name: clusterName, Ip: "2.3.4.5", Port: 1234}
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.eventsChannel <- event
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.topologyRequestChannel <- request
}

func (m *SentinelManager) ScheduleNewHealthChecker(sentinel types.Sentinel) {
	logger.Info.Printf("SentinelManager : scheduling new healthChecker for , %s", util.String(sentinel))
	util.Schedule(func() { m.NewSentinelMonitor(sentinel) }, time.Second*5)
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

		m.ScheduleNewHealthChecker(sentinel)
		logger.Trace.Printf("Sentinel lost : %s (scheduling new health checker).", util.String(topologyState))

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
