package sentinel

import (
	"strconv"
	"strings"
	"time"

	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
)

type Monitor struct {
	pubSubClient    *redis.PubSubClient
	client          *SentinelClient
	channel         chan redis.RedisPubSubReply
	manager         Manager
	sentinel        types.Sentinel
	redisConnection redis.RedisConnection
}

func NewMonitor(sentinel types.Sentinel, manager Manager, redisConnection redis.RedisConnection) (*Monitor, error) {

	uri := sentinel.GetLocation()

	channel := make(chan redis.RedisPubSubReply)
	pubSubClient, err := redis.NewPubSubClient(uri, channel, redisConnection)

	if err != nil {
		return nil, err
	}

	client, err := NewSentinelClient(sentinel, redisConnection)

	if err != nil {
		return nil, err
	}

	monitor := &Monitor{pubSubClient: pubSubClient,
		client:          client,
		channel:         channel,
		manager:         manager,
		sentinel:        sentinel,
		redisConnection: redisConnection}
	return monitor, nil
}

func (m *Monitor) StartMonitoringMasterEvents(switchmasterchannel chan types.MasterSwitchedEvent) error {

	keys := []string{"+switch-master", "+sentinel"}
	err := m.pubSubClient.Start(keys)

	if err != nil {
		logger.Error.Printf("Error StartMonitoringMasterEvents %s sentinel@%s", err.Error(), m.sentinel.GetLocation())
		m.pubSubClient.Close()
		return err
	}

	go m.loop(switchmasterchannel)

	return nil
}

func (m *Monitor) loop(switchmasterchannel chan types.MasterSwitchedEvent) {
L:
	for {
		select {
		case message := <-m.channel:
			shutdown := m.dealWithSentinelMessage(message, switchmasterchannel)
			if shutdown {

				m.manager.Notify(&SentinelLost{Sentinel: m.sentinel})
				logger.Info.Printf("Shutting down monitor %s", m.sentinel.GetLocation())

				break L
			}

		case <-time.After(MonitorPingInterval):

			resp := m.client.Ping()
			if resp != "PONG" {

				logger.Info.Printf("Error pinging client : %s, resonse : %s", m.sentinel.GetLocation(), resp)
				m.manager.Notify(&SentinelLost{Sentinel: m.sentinel})
				logger.Info.Printf("Shutting down monitor %s", m.sentinel.GetLocation())

				break L
			}

			m.manager.Notify(&SentinelPing{Sentinel: m.sentinel})

			knownClusters := m.client.FindKnownClusters()

			m.manager.Notify(&SentinelClustersMonitoredUpdate{Sentinel: m.sentinel, Clusters: knownClusters})

			for _, clustername := range knownClusters {

				sentinels := m.client.FindConnectedSentinels(clustername)

				for _, connectedsentinel := range sentinels {
					m.manager.Notify(&SentinelAdded{Sentinel: connectedsentinel})
				}
			}
		}
	}
}

func (m *Monitor) dealWithSentinelMessage(message redis.RedisPubSubReply, switchmasterchannel chan types.MasterSwitchedEvent) bool {

	if message.Timeout() {
		return true
	}
	if message.Err() != nil {
		logger.Info.Printf("Subscription Message : Channel : Error %s", message.Err())
		return true
	}

	channel := message.Channel()

	if channel == "+switch-master" {
		logger.Info.Printf("Subscription Message : Channel : %s : %s", message.Channel(), message.Message())

		event := parseSwitchMasterMessage(message.Message())
		switchmasterchannel <- event
		return false
	}
	if channel == "+sentinel" {

		host, port := parseInstanceDetailsForIpAndPortMessage(message.Message())
		m.manager.Notify(&SentinelAdded{Sentinel: types.Sentinel{Host: host, Port: port}})
		return false
	}
	return false
}

func parseInstanceDetailsForIpAndPortMessage(message string) (string, int) {
	//<instance-type> <name> <ip> <port> @ <master-name> <master-ip> <master-port>
	bits := strings.Split(message, " ")
	port, _ := strconv.Atoi(bits[3])
	return bits[2], port
}

func parseSwitchMasterMessage(message string) types.MasterSwitchedEvent {
	bits := strings.Split(message, " ")

	oldmasterport, _ := strconv.Atoi(bits[2])
	newmasterport, _ := strconv.Atoi(bits[4])

	return types.MasterSwitchedEvent{Name: bits[0], OldMasterIp: bits[1], OldMasterPort: oldmasterport, NewMasterIp: bits[3], NewMasterPort: newmasterport}
}
