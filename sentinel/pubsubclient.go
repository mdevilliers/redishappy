package sentinel

import (
	"github.com/fzzy/radix/extra/pubsub"
	"github.com/fzzy/radix/redis"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"strconv"
	"strings"
)

type SentinelPubSubClient struct {
	subscriptionclient *pubsub.SubClient
	redisclient        *redis.Client
}

type MasterSwitchedEvent struct {
	Name          string
	OldMasterIp   string
	OldMasterPort int
	NewMasterIp   string
	NewMasterPort int
}

func NewPubSubClient(sentinel types.Sentinel) (*SentinelPubSubClient, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("Connecting to sentinel@%s", uri)

	redisclient, err := redis.Dial("tcp", uri)

	if err != nil {
		logger.Error.Printf("Error connecting to sentinel@%s", uri, err.Error())
		return nil, err
	}

	logger.Info.Printf("Connected to sentinel@%s", uri)

	redissubscriptionclient := pubsub.NewSubClient(redisclient)

	client := &SentinelPubSubClient{redisclient: redisclient,
		subscriptionclient: redissubscriptionclient}
	return client, nil
}

func (client *SentinelPubSubClient) StartMonitoringMasterEvents(switchmasterchannel chan MasterSwitchedEvent) error {

	//TODO : fix radix client - doesn't support PSubscribe
	subr := client.subscriptionclient.Subscribe("+switch-master") //, "+slave-reconf-done ")

	if subr.Err != nil {
		return subr.Err
	}

	go client.loopSubscription(switchmasterchannel)

	return nil
}

func (sub *SentinelPubSubClient) loopSubscription(switchmasterchannel chan MasterSwitchedEvent) {
	for {
		r := sub.subscriptionclient.Receive()
		if r.Timeout() {
			continue
		}
		if r.Err == nil {
			logger.Info.Printf("Subscription Message : Channel : %s : %s\n", r.Channel, r.Message)

			if r.Channel == "+switch-master" {
				bits := strings.Split(r.Message, " ")

				oldmasterport, _ := strconv.Atoi(bits[2])
				newmasterport, _ := strconv.Atoi(bits[4])

				event := MasterSwitchedEvent{Name: bits[0], OldMasterIp: bits[1], OldMasterPort: oldmasterport, NewMasterIp: bits[3], NewMasterPort: newmasterport}
				switchmasterchannel <- event
			}
		} else {
			logger.Info.Printf("Subscription Message : Channel : Error %s \n", r.Err)
		}
	}
}
