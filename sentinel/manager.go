package sentinel

import (
	"time"

	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
)

const (
	SentinelMarkedUp      = 1
	SentinelMarkedDown    = 2
	SentinelMarkedAlive   = 3
	SentinelMarkedUnknown = 4
)

const (
	SentinelReconnectionPeriod = time.Second * 5
	MonitorPingInterval        = time.Second * 1
)

type Manager interface {
	Notify(event SentinelEvent)
}

type SentinelManager struct {
	switchmasterchannel  chan types.MasterSwitchedEvent
	connectionChannel    chan types.ConnectionEvent
	redisConnection      redis.RedisConnection
	configurationManager *configuration.ConfigurationManager
	throttle             *Throttle
	state                SentinelState
}

func NewManager(switchmasterchannel chan types.MasterSwitchedEvent, cm *configuration.ConfigurationManager) *SentinelManager {

	unthrottled := make(chan types.MasterSwitchedEvent)
	throttle := NewThrottle(unthrottled, switchmasterchannel)

	connectionChannel := make(chan types.ConnectionEvent)

	manager := &SentinelManager{
		switchmasterchannel:  unthrottled,
		connectionChannel:    connectionChannel,
		redisConnection:      redis.RedisConnection{},
		configurationManager: cm,
		throttle:             throttle,
	}

	startMonitoringCallback := func(sentinel types.Sentinel) {

		manager.Notify(&SentinelUnknown{Sentinel: sentinel})
		go manager.startNewMonitor(sentinel)
	}

	manager.state = NewSentinelState(startMonitoringCallback)

	go manager.bootstrap()
	return manager
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.state.Notify(event)
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.state.GetState(request)
}

func (m *SentinelManager) GetCurrentTopology() types.MasterDetailsCollection {
	stateChannel := make(chan types.MasterDetailsCollection)
	go m.getTopology(stateChannel)
	return <-stateChannel
}

func (m *SentinelManager) startNewMonitor(sentinel types.Sentinel) {

	monitor, err := NewMonitor(sentinel, m, m.redisConnection, m.configurationManager.GetCurrentConfiguration().SentinelTCPKeepAlive)

	if err != nil {
		logger.Error.Printf("Error starting monitor %s : %s", sentinel.GetLocation(), err.Error())
		m.Notify(&SentinelLost{Sentinel: sentinel})
		return
	}

	err = monitor.StartMonitoringMasterEvents(m.switchmasterchannel, m.connectionChannel)

	if err != nil {
		logger.Error.Printf("Error starting monitoring events %s : %s", sentinel.GetLocation(), err.Error())
		m.Notify(&SentinelLost{Sentinel: sentinel})
	}

	// start our connection channel reciever?
	go m.receiveConnectionMessage(m.connectionChannel, m.switchmasterchannel)

}

func (m *SentinelManager) receiveConnectionMessage(in chan types.ConnectionEvent, out chan types.MasterSwitchedEvent) {
	for {
		select {
		case event := <-in:
			if event.Connected {

				// types.MasterDetailsCollection{items:map[string]*types.MasterDetails{"redis-default":(*types.MasterDetails)(0xc8201f19b0)
				currentTopology := m.GetCurrentTopology()
				logger.Trace.Printf("%#v\n", currentTopology)

				for _, v := range currentTopology.Items() {
					// Tpoplog: types.MasterDetails{ExternalPort:6379, Name:"redis-default", Ip:"10.10.10.101", Port:6379}
					logger.Info.Printf("Resync Topology: %#v\n", *v)

					resyncMasterSwitchEvent := types.MasterSwitchedEvent{
						Name:          v.Name,
						OldMasterIp:   "resync",
						OldMasterPort: 0,
						NewMasterIp:   v.Ip,
						NewMasterPort: v.Port,
					}

					out <- resyncMasterSwitchEvent
				}
			}
		}
	}
}

func (m *SentinelManager) getTopology(stateChannel chan types.MasterDetailsCollection) {

	topology := types.NewMasterDetailsCollection()
	configuration := m.configurationManager.GetCurrentConfiguration()

	for _, sentinel := range configuration.Sentinels {
		client, err := redis.NewSentinelClient(sentinel, m.redisConnection, configuration.SentinelTCPKeepAlive)

		if err != nil {
			logger.Error.Printf("Error starting sentinel (%s) client : %s", sentinel.GetLocation(), err.Error())
			continue
		}
		defer client.Close()

		for _, clusterDetails := range configuration.Clusters {

			details, err := client.DiscoverMasterForCluster(clusterDetails.Name)

			if err != nil {
				continue
			}

			details.ExternalPort = clusterDetails.ExternalPort
			// TODO : last one wins?
			topology.AddOrReplace(details)
		}
	}
	stateChannel <- topology
}

func (m *SentinelManager) bootstrap() {

	configuration := m.configurationManager.GetCurrentConfiguration()

	for _, sentinel := range configuration.Sentinels {
		m.Notify(&SentinelAdded{Sentinel: sentinel})
	}
}
