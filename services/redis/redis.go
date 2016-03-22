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
	Message() string
	Channel() string
	MessageType() int
}

const (
	Confirmation = 1
	Message      = 2
)

type PubSubReply struct {
	message     string
	channel     string
	err         error
	messageType int
}

func NewRedisPubSubReply(message []string, err error) RedisPubSubReply {

	if err != nil {
		return &PubSubReply{err: err}
	}

	ret := &PubSubReply{
		channel: message[1],
	}

	if message[0] == "subscribe" {
		ret.messageType = Confirmation
	}

	if message[0] == "message" {
		ret.messageType = Message
	}

	ret.message = message[2]
	return ret
}

func (p PubSubReply) Err() error {
	return p.err
}
func (p PubSubReply) Message() string {
	return p.message
}

func (p PubSubReply) Channel() string {
	return p.channel
}

func (p PubSubReply) MessageType() int {
	return p.messageType
}
