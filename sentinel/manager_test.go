package sentinel

import (
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"testing"
)

func TestBasicEventChannel(t *testing.T) {
	logger.InitLogging("../log")
	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel, &configuration.Configuration{})
	defer manager.ClearState()
	manager.Notify(&SentinelAdded{Sentinel: types.Sentinel{Host: "10.1.1.1", Port: 12345}})

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	manager2 := NewManager(switchmasterchannel, &configuration.Configuration{})
	manager2.Notify(&SentinelAdded{Sentinel: types.Sentinel{Host: "10.1.1.2", Port: 12345}})

	manager2.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState = <-responseChannel

	if len(topologyState.Sentinels) != 2 {
		t.Errorf("Topology count should be 2 : it is %d", len(topologyState.Sentinels))
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAddingAndLoseingASentinel(t *testing.T) {
	logger.InitLogging("../log")
	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel, &configuration.Configuration{})
	defer manager.ClearState()

	sentinel := types.Sentinel{Host: "10.1.1.5", Port: 12345}

	manager.Notify(&SentinelAdded{Sentinel: sentinel})
	manager.Notify(&SentinelLost{Sentinel: sentinel})

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAddingSentinelMultipleTimes(t *testing.T) {
	logger.InitLogging("../log")
	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel, &configuration.Configuration{})
	defer manager.ClearState()

	sentinel := types.Sentinel{Host: "10.1.1.6", Port: 12345}

	manager.Notify(&SentinelAdded{Sentinel: sentinel})

	ping := &SentinelPing{Sentinel: sentinel}
	ping2 := &SentinelPing{Sentinel: sentinel}
	manager.Notify(ping)
	manager.Notify(ping2)

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	_, ok := topologyState.FindSentinelInfo(sentinel)

	if !ok {
		t.Error("Added sentinel not found")
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAllSentinelsFromTheConfigurationAreAddedToTheTopology(t *testing.T) {
	logger.InitLogging("../log")

	sentinels := []types.Sentinel{types.Sentinel{Host: "1.2.3.4"}, types.Sentinel{Host: "2.3.4.5"}}
	config := &configuration.Configuration{Sentinels: sentinels}

	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel, config)
	defer manager.ClearState()

	responseChannel := make(chan SentinelTopology)
	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 2 {
		t.Errorf("Two sentinels should have been added %d", len(topologyState.Sentinels))
	}

}
