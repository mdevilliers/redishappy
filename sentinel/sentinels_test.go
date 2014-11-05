package sentinel

import (
	"fmt"
	"testing"

	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

func TestCanAddSentinels(t *testing.T) {
	state := NewSentinelState(func(_ types.Sentinel) {})
	sentinel := types.Sentinel{Host: "10.1.1.2", Port: 12345}
	sentinelMessage := SentinelAdded{Sentinel: sentinel}
	state.Notify(&sentinelMessage)

	responseChannel := make(chan SentinelTopology)
	state.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	sen, _ := topologyState.FindSentinelInfo(sentinel)

	if sen.SentinelLocation != sentinel.GetLocation() {
		t.Error("Wrong host found")
	}

	if sen.State != SentinelMarkedUp {
		t.Error("Sentinel in wrong state")
	}
}

func TestCanMarkSentinelsDown(t *testing.T) {

	state := NewSentinelState(func(_ types.Sentinel) {})
	sentinel := types.Sentinel{Host: "10.1.1.2", Port: 12345}

	sentinelAddedMessage := &SentinelAdded{Sentinel: sentinel}
	sentinelLostMessage := &SentinelLost{Sentinel: sentinel}

	responseChannel := make(chan SentinelTopology)

	state.Notify(sentinelAddedMessage)
	state.Notify(sentinelLostMessage)

	state.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel
	fmt.Print(util.String(topologyState))

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	sen, _ := topologyState.FindSentinelInfo(sentinel)

	if sen.SentinelLocation != sentinel.GetLocation() {
		t.Error("Wrong host found")
	}

	if sen.State != SentinelMarkedDown {
		t.Errorf("Sentinel in wrong state : %d", sen.State)
	}
}

func TestCanMarkSentinelsAlive(t *testing.T) {

	state := NewSentinelState(func(_ types.Sentinel) {})
	sentinel := types.Sentinel{Host: "10.1.1.2", Port: 12345}
	sentinelMessage := SentinelAdded{Sentinel: sentinel}
	sentinelPingMessage := SentinelPing{Sentinel: sentinel}
	state.Notify(&sentinelMessage)
	state.Notify(&sentinelPingMessage)

	responseChannel := make(chan SentinelTopology)
	state.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	sen, _ := topologyState.FindSentinelInfo(sentinel)

	if sen.SentinelLocation != sentinel.GetLocation() {
		t.Error("Wrong host found")
	}

	if sen.State != SentinelMarkedAlive {
		t.Error("Sentinel in wrong state")
	}
}

func TestCanAddCLustersToSentinel(t *testing.T) {

	state := NewSentinelState(func(_ types.Sentinel) {})
	sentinel := types.Sentinel{Host: "10.1.1.2", Port: 12345}

	sentinelMessage := SentinelAdded{Sentinel: sentinel}
	sentinelInfoMessage := SentinelClustersMonitoredUpdate{Sentinel: sentinel, Clusters: []string{"A"}}

	state.Notify(&sentinelMessage)
	state.Notify(&sentinelInfoMessage)

	responseChannel := make(chan SentinelTopology)
	state.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	sen, _ := topologyState.FindSentinelInfo(sentinel)

	if sen.SentinelLocation != sentinel.GetLocation() {
		t.Error("Wrong host found")
	}

	if len(sen.Clusters) != 1 {
		t.Error("Wrong number of clusters")
	}

	if sen.Clusters[0] != "A" {
		t.Error("Wrong cluster found")
	}
}

func TestUnknowSentinelsAreNoOps(t *testing.T) {

	state := NewSentinelState(func(_ types.Sentinel) {})
	sentinel := types.Sentinel{Host: "10.1.1.3", Port: 12345}

	state.Notify(&SentinelClustersMonitoredUpdate{Sentinel: sentinel, Clusters: []string{"A"}})
	state.Notify(&SentinelLost{Sentinel: sentinel})
	state.Notify(&SentinelPing{Sentinel: sentinel})
	state.Notify(&SentinelUnknown{Sentinel: sentinel})

	responseChannel := make(chan SentinelTopology)
	state.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 0 {
		t.Error("No sentinel was added so the state should be empty")
	}
}
