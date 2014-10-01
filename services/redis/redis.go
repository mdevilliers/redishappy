package redis

import (
	"github.com/fzzy/radix/extra/pubsub"
	"github.com/fzzy/radix/redis"
)

type RedisConnection interface {
	GetConnection(protocol, uri string) (RedisClient, error)
}

type RedisReply interface {
	List() ([]string, error)
	String() string
	Err() error
	Elems() []RedisReply
	Hash() (map[string]string, error)
}

type RedisClient interface {
	Cmd(cmd string, args ...interface{}) RedisReply
	NewPubSubClient() RedisPubSubClient
}

type RedisPubSubClient interface {
	Subscribe(channels ...interface{}) RedisPubSubReply
	Receive() RedisPubSubReply
}

type RedisPubSubReply interface {
	Err() error
	Timeout() bool
	Channel() string
	Message() string
}

type RadixRedisConnection struct {
}

type RadixRedisClient struct {
	client *redis.Client
}

type RadixRedisReply struct {
	reply *redis.Reply
}

type RadixPubSubClient struct {
	client *pubsub.SubClient
}

type RadixPubSubReply struct {
	reply *pubsub.SubReply
}

func (c RadixRedisConnection) GetConnection(protocol, uri string) (RedisClient, error) {
	redisclient, err := redis.Dial(protocol, uri)
	return &RadixRedisClient{client: redisclient}, err
}

func (c *RadixRedisClient) Cmd(cmd string, args ...interface{}) RedisReply {
	re := c.client.Cmd(cmd, args)
	return makeRedisReply(re)
}

func (c *RadixRedisReply) String() string {
	return c.reply.String()
}

func (c *RadixRedisReply) Err() error {
	return c.reply.Err
}

func (c *RadixRedisReply) Elems() []RedisReply {
	elements := c.reply.Elems
	reply := make([]RedisReply, len(elements))

	for n, element := range elements {
		reply[n] = makeRedisReply(element)
	}

	return reply
}
func makeRedisReply(re *redis.Reply) RedisReply {
	return &RadixRedisReply{reply: re}
}

func (c *RadixRedisReply) Hash() (map[string]string, error) {
	return c.reply.Hash()
}

func (c *RadixRedisReply) List() ([]string, error) {
	return c.reply.List()
}

func (c *RadixRedisClient) NewPubSubClient() RedisPubSubClient {
	client := pubsub.NewSubClient(c.client)
	return &RadixPubSubClient{client: client}
}

func (c *RadixPubSubClient) Subscribe(channels ...interface{}) RedisPubSubReply {
	reply := c.client.Subscribe(channels)
	return &RadixPubSubReply{reply: reply}
}
func (c *RadixPubSubClient) Receive() RedisPubSubReply {
	reply := c.client.Receive()
	return &RadixPubSubReply{reply: reply}
}

func (r *RadixPubSubReply) Err() error {
	return r.reply.Err
}

func (r *RadixPubSubReply) Timeout() bool {
	return r.reply.Timeout()
}

func (r *RadixPubSubReply) Channel() string {
	return r.reply.Channel
}

func (r *RadixPubSubReply) Message() string {
	return r.reply.Message
}
