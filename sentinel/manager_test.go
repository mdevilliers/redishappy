package sentinel

import (
	// "fmt"
	"testing"
	"time"
	"github.com/mdevilliers/redishappy/configuration"
)

func TestBasicEventChannel(t *testing.T) {

	manager := NewManager()

	manager.Notify(&SentinelAdded{ sentinel : &configuration.Sentinel {Host : "10.1.1.1", Port : 12345}})

	responseChannel := make (chan SentinelTopology)

	go func() {
		select {
	            case topologyState := <-responseChannel:
	                if topologyState.Count != 1 {
       					t.Error("Topology count should be 1")
       				}
            }
    }()

	manager.GetState(TopologyRequest{ReplyChannel : responseChannel})
	// TODO : get rid of this
	time.Sleep(100 * time.Millisecond)

	manager2 := NewManager()
	manager2.Notify(&SentinelAdded{ sentinel : &configuration.Sentinel {Host : "10.1.1.2", Port : 12345}})

	go func() {
		select {
	            case topologyState := <-responseChannel:
	                if topologyState.Count != 2 {
       					t.Errorf("Topology count should be 2 : it is %d", topologyState.Count)
       				}
            }
    }()

	manager2.GetState(TopologyRequest{ReplyChannel : responseChannel})
	// TODO : get rid of this
	time.Sleep(100 * time.Millisecond)

}
