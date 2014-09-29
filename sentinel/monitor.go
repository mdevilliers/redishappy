package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"strconv"
	"strings"
)

type Monitor struct {
	client  *redis.PubSubClient
	channel chan redis.RedisPubSubReply
}

func NewMonitor(sentinel types.Sentinel, redisConnection redis.RedisConnection) (*Monitor, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("Connecting to sentinel@%s", uri)

	channel := make(chan redis.RedisPubSubReply)
	key := "+switch-master" //, "+slave-reconf-done ")

	client, err := redis.NewPubSubClient(uri, key, channel, redisConnection)

	if err != nil {
		return nil, err
	}

	monitor := &Monitor{client: client, channel: channel}
	return monitor, nil
}

func (m *Monitor) StartMonitoringMasterEvents(switchmasterchannel chan types.MasterSwitchedEvent) error {

	err := m.client.Start()

	if err != nil {
		return err
	}

	go m.loop(switchmasterchannel)

	return nil
}

func (m *Monitor) loop(switchmasterchannel chan types.MasterSwitchedEvent) {
	for {
		message := <-m.channel

		if message.Timeout() {
			continue
		}
		if message.Err() == nil {
			logger.Info.Printf("Subscription Message : Channel : %s : %s\n", message.Channel, message.Message)

			event := parseSwitchMasterMessage(message.Message())
			switchmasterchannel <- event

		} else {
			logger.Info.Printf("Subscription Message : Channel : Error %s \n", message.Err)
			break
		}
	}
}

func parseSwitchMasterMessage(message string) types.MasterSwitchedEvent {
	bits := strings.Split(message, " ")

	oldmasterport, _ := strconv.Atoi(bits[2])
	newmasterport, _ := strconv.Atoi(bits[4])

	return types.MasterSwitchedEvent{Name: bits[0], OldMasterIp: bits[1], OldMasterPort: oldmasterport, NewMasterIp: bits[3], NewMasterPort: newmasterport}
}
