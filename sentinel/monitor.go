package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"strconv"
	"strings"
	"time"
)

type Monitor struct {
	client   *redis.PubSubClient
	channel  chan redis.RedisPubSubReply
	manager  Manager
	sentinel types.Sentinel
}

func NewMonitor(sentinel types.Sentinel, manager Manager, redisConnection redis.RedisConnection) (*Monitor, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("Connecting to sentinel@%s", uri)

	channel := make(chan redis.RedisPubSubReply)
	client, err := redis.NewPubSubClient(uri, channel, redisConnection)

	if err != nil {
		return nil, err
	}

	monitor := &Monitor{client: client, channel: channel, manager: manager, sentinel: sentinel}
	return monitor, nil
}

func (m *Monitor) StartMonitoringMasterEvents(switchmasterchannel chan types.MasterSwitchedEvent) error {

	keys := []string{"+switch-master", "+sentinel"}
	err := m.client.Start(keys)

	if err != nil {
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
				logger.Info.Printf("Shutting down monitor %s", m.sentinel)
				break L
			}

		case <-time.After(time.Duration(1) * time.Second):
			m.manager.Notify(&SentinelPing{Sentinel: m.sentinel})
		}
	}
}

func (m *Monitor) dealWithSentinelMessage(message redis.RedisPubSubReply, switchmasterchannel chan types.MasterSwitchedEvent) bool {

	if message.Timeout() {
		return true
	}
	if message.Err() != nil {
		m.manager.Notify(&SentinelLost{Sentinel: m.sentinel})
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
