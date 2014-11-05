package redis

import (
	"github.com/mdevilliers/redishappy/services/logger"
)

type PubSubClient struct {
	subscriptionClient RedisPubSubClient
	channel            chan RedisPubSubReply
}

func NewPubSubClient(url string, channel chan RedisPubSubReply, redisConnection RedisConnection) (*PubSubClient, error) {

	client, err := redisConnection.GetConnection("tcp", url)

	if err != nil {
		logger.Error.Printf("PubSubClient Error connecting to %s : %s", url, err.Error())
		return nil, err
	}

	subclient := &PubSubClient{subscriptionClient: client.NewPubSubClient(), channel: channel}
	return subclient, nil
}

func (client *PubSubClient) Start(keys []string) error {

	subr := client.subscriptionClient.Subscribe(keys)

	if subr.Err() != nil {
		return subr.Err()
	}

	go client.loopSubscription()

	return nil
}

func (client *PubSubClient) Close() {
	client.subscriptionClient.Close()
}

func (client *PubSubClient) loopSubscription() {
	for {
		r := client.subscriptionClient.Receive()
		client.channel <- r
	}
}
