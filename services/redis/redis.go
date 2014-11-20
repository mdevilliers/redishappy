package redis

import (
	"time"

	"github.com/therealbill/libredis/client"
)

const (
	RedisConnectionTimeoutPeriod = time.Second * 2
)

type RedisConnection struct{}

type Redis interface {
	GetConnection(protocol, uri string) (RedisClient, error)
}

func (RedisConnection) GetConnection(protocol, uri string) (RedisClient, error) {
	client, err := client.DialWithConfig(&client.DialConfig{
		Network: protocol,
		Address: uri,
		Timeout: RedisConnectionTimeoutPeriod})
	return client, err
}

type RedisClient interface {
	ClosePool()
	SentinelGetMaster(cluster string) (client.MasterAddress, error)
	SentinelSentinels(cluster string) ([]client.SentinelInfo, error)
	SentinelMasters() ([]client.MasterInfo, error)
	Ping() error
	PubSub() (*client.PubSub, error)
}

type RedisPubSubReply interface {
	Err() error
	Message() []string
}

type PubSubReply struct {
	message []string
	err     error
}

func (p PubSubReply) Err() error {
	return p.err
}
func (p PubSubReply) Message() []string {
	return p.message
}
