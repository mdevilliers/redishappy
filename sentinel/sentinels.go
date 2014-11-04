package sentinel

import (
	"time"

	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type StartMonitoringSentinel func(types.Sentinel)

type SentinelState struct {
	notifyChannel           chan SentinelEvent
	readChannel             chan TopologyRequest
	state                   SentinelTopology
	startMonitoringSentinel StartMonitoringSentinel
}

func NewSentinelState(fn StartMonitoringSentinel) SentinelState {
	events := make(chan SentinelEvent)
	requests := make(chan TopologyRequest)

	agent := SentinelState{
		state:                   SentinelTopology{Sentinels: map[string]*SentinelInfo{}},
		notifyChannel:           events,
		readChannel:             requests,
		startMonitoringSentinel: fn,
	}

	go agent.loopEvents(events, requests)

	return agent
}

func (s SentinelState) Notify(event SentinelEvent) {
	s.notifyChannel <- event
}

func (s SentinelState) GetState(request TopologyRequest) {
	s.readChannel <- request
}

func (s SentinelState) loopEvents(events chan SentinelEvent, topology chan TopologyRequest) {
	for {
		select {
		case event := <-events:
			s.updateState(event)
		case read := <-topology:
			read.ReplyChannel <- s.state
		}
	}
}

func (s SentinelState) updateState(event interface{}) {

	switch e := event.(type) {
	case *SentinelAdded:

		sentinel := e.GetSentinel()
		uid := s.state.createKey(sentinel)

		//if we don't know about the sentinel start monitoring it
		if _, exists := s.state.Sentinels[uid]; !exists {

			info := &SentinelInfo{SentinelLocation: uid,
				LastUpdated: time.Now().UTC(),
				State:       SentinelMarkedUp}

			s.state.Sentinels[uid] = info

			go s.startMonitoringSentinel(sentinel)

			logger.Trace.Printf("Sentinel added : %s", util.String(s.state))
		}

	case *SentinelLost:

		sentinel := e.GetSentinel()
		uid := s.state.createKey(sentinel)
		currentInfo, ok := s.state.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedDown
			currentInfo.LastUpdated = time.Now().UTC()

			util.Schedule(func() { go s.startMonitoringSentinel(sentinel) }, SentinelReconnectionPeriod)

			logger.Trace.Printf("Sentinel lost : %s (scheduling new client and monitor).", util.String(s.state))

		} else {
			logger.Trace.Printf("Unknown sentinel lost : %s.", util.String(s.state))
		}

	case *SentinelPing:
		sentinel := e.GetSentinel()
		uid := s.state.createKey(sentinel)
		currentInfo, ok := s.state.Sentinels[uid]

		if ok {
			currentInfo.State = SentinelMarkedAlive
			currentInfo.LastUpdated = time.Now().UTC()
		} else {
			logger.Trace.Printf("Unknown sentinel ping : %s.", util.String(s.state))
		}

	case *SentinelClustersMonitoredUpdate:
		sentinel := e.GetSentinel()
		uid := s.state.createKey(sentinel)

		if info, exists := s.state.Sentinels[uid]; exists {

			info.Clusters = e.Clusters
		} else {
			logger.Trace.Printf("Unknown sentinel updated state : %s.", util.String(s.state))
		}

	default:
		logger.Error.Println("Unknown sentinel event : ", util.String(e))
	}
}
