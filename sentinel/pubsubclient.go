package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"strconv"
	"strings"
)

type SentinelPubSubClient struct {
	subscriptionClient redis.RedisPubSubClient
}

type MasterSwitchedEvent struct {
	Name          string
	OldMasterIp   string
	OldMasterPort int
	NewMasterIp   string
	NewMasterPort int
}

func NewPubSubClient(sentinel types.Sentinel, redisConnection redis.RedisConnection) (*SentinelPubSubClient, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("Connecting to sentinel@%s", uri)

	redisclient, err := redisConnection.Dial("tcp", uri)

	if err != nil {
		logger.Error.Printf("Error connecting to sentinel@%s", uri, err.Error())
		return nil, err
	}

	logger.Info.Printf("Connected to sentinel@%s", uri)

	redissubscriptionclient := redisclient.NewPubSubClient()

	client := &SentinelPubSubClient{subscriptionClient: redissubscriptionclient}
	return client, nil
}

func (client *SentinelPubSubClient) StartMonitoringMasterEvents(switchmasterchannel chan MasterSwitchedEvent) error {

	subr := client.subscriptionClient.Subscribe("+switch-master") //, "+slave-reconf-done ")

	if subr.Err() != nil {
		return subr.Err()
	}

	go client.loopSubscription(switchmasterchannel)

	return nil
}

func (sub *SentinelPubSubClient) loopSubscription(switchmasterchannel chan MasterSwitchedEvent) {
	for {
		r := sub.subscriptionClient.Receive()
		if r.Timeout() {
			continue
		}
		if r.Err() == nil {
			logger.Info.Printf("Subscription Message : Channel : %s : %s\n", r.Channel, r.Message)

			if r.Channel() == "+switch-master" {
				event := parseSwitchMasterMessage(r.Message())
				switchmasterchannel <- event
			}
		} else {
			logger.Info.Printf("Subscription Message : Channel : Error %s \n", r.Err)
			break
		}
	}
}

func parseSwitchMasterMessage(message string) MasterSwitchedEvent {
	bits := strings.Split(message, " ")

	oldmasterport, _ := strconv.Atoi(bits[2])
	newmasterport, _ := strconv.Atoi(bits[4])

	return MasterSwitchedEvent{Name: bits[0], OldMasterIp: bits[1], OldMasterPort: oldmasterport, NewMasterIp: bits[3], NewMasterPort: newmasterport}
}
