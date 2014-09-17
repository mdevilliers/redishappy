package sentinel

import (
	"github.com/mdevilliers/redishappy/types"
	"testing"
)

func TestBasicEventChannel(t *testing.T) {

	manager := NewManager()
	defer manager.ClearState()
	manager.Notify(&SentinelAdded{sentinel: &types.Sentinel{Host: "10.1.1.1", Port: 12345}})

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	manager2 := NewManager()
	manager2.Notify(&SentinelAdded{sentinel: &types.Sentinel{Host: "10.1.1.2", Port: 12345}})

	manager2.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState = <-responseChannel

	if len(topologyState.Sentinels) != 2 {
		t.Errorf("Topology count should be 2 : it is %d", len(topologyState.Sentinels))
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAddingAndLoseingASentinel(t *testing.T) {

	manager := NewManager()
	defer manager.ClearState()

	sentinel := &types.Sentinel{Host: "10.1.1.5", Port: 12345}

	manager.Notify(&SentinelAdded{sentinel: sentinel})
	manager.Notify(&SentinelLost{sentinel: sentinel})

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAddingInfoToADiscoveredSentinel(t *testing.T) {

	manager := NewManager()
	defer manager.ClearState()

	sentinel := &types.Sentinel{Host: "10.1.1.6", Port: 12345}

	manager.Notify(&SentinelAdded{sentinel: sentinel})

	ping := &SentinelPing{sentinel: sentinel, Clusters: []string{"one", "two", "three"}}
	ping2 := &SentinelPing{sentinel: sentinel, Clusters: []string{"four", "five"}}
	manager.Notify(ping)
	manager.Notify(ping2)

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	info, ok := topologyState.FindSentinelInfo(sentinel)

	if ok {
		if len(info.KnownClusters) != 2 {
			t.Error("There should only be 2 known clusters")
		}
	} else {
		t.Error("Added sentinel not found")
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}
