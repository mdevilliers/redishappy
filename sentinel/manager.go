package sentinel

import (
	"fmt"
	"github.com/mdevilliers/redishappy/configuration"
	"sync"
)


type SentinelTopology struct{
	Count int
}

type TopologyRequest struct{
	ReplyChannel chan SentinelTopology
}

type SentinelEvent interface{
	GetHost() string
	GetPort() int
}

type SentinelAdded struct{
	sentinel *configuration.Sentinel
}
// type SentinelLost struct{
// 	c *configuration.Sentinel
// }
// type SetinelInfo struct{
// 	c *configuration.Sentinel
// 	Clusters []string
// }
// type SentinelDiscovered struct {
//  c *configuration.Sentinel
// }

type SentinelManager struct {
	eventsChannel chan SentinelEvent
	topologyRequestChannel chan TopologyRequest
}

func( s SentinelAdded) GetHost() string {
	return s.sentinel.Host
}
func( s SentinelAdded) GetPort() int {
	return s.sentinel.Port
}
// func( s SentinelAdded) GetHost() string {
// 	return s.sentinel.Host
// }
// func( s SentinelAdded) GetPort() int {
// 	return s.sentinel.Port
// }


var topologyState = SentinelTopology{Count : 0}
var statelock = &sync.Mutex{}

func NewManager() *SentinelManager {
	events := make (chan SentinelEvent)
	requests := make (chan TopologyRequest)
	go loopEvents(events, requests)
	return &SentinelManager{ eventsChannel : events, topologyRequestChannel:requests}
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.eventsChannel <- event
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.topologyRequestChannel <- request
}

func loopEvents(events chan SentinelEvent, topology chan TopologyRequest) {

 	for {
            select {
	            case event := <- events:
	            	statelock.Lock()
	                fmt.Printf("Got an event : %s : %d\n", event.GetHost(), event.GetPort())
					topologyState.Count ++	
					fmt.Printf("Updated CurrentState : Count %d\n", topologyState.Count )
					statelock.Unlock()
	            case read := <-topology:
	            	fmt.Printf("Sending CurrentState : Count %d\n", topologyState.Count )
	                read.ReplyChannel <- topologyState	            
            }
        }
}