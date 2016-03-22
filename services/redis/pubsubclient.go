package redis

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/therealbill/libredis/client"
)

type PubSubClient struct {
	subscriptionClient *client.PubSub
	channel            chan RedisPubSubReply
}

func NewPubSubClient(url string, channel chan RedisPubSubReply, redisConnection RedisConnection, tcp_keepalive int) (*PubSubClient, error) {

	client, err := redisConnection.GetConnection("tcp", url, tcp_keepalive)

	if err != nil {
		logger.Error.Printf("PubSubClient Error connecting to %s : %s", url, err.Error())
		return nil, err
	}
	pubsub, err := client.PubSub()

	if err != nil {
		logger.Error.Printf("Error creating pub sub client : %s : %s", url, err.Error())
		return nil, err
	}
	subclient := &PubSubClient{subscriptionClient: pubsub, channel: channel}
	return subclient, nil
}

func (client *PubSubClient) Start(key string) error {

	err := client.subscriptionClient.Subscribe(key)

	if err != nil {
		return err
	}

	go client.loopSubscription()

	return nil
}

func (client *PubSubClient) Close() {
	client.subscriptionClient.Close()
}

func (client *PubSubClient) loopSubscription() {
	for {
		r, err := client.subscriptionClient.Receive()
		client.channel <- NewRedisPubSubReply(r, err)
	}
}
