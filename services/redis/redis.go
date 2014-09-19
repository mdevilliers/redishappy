package redis

import(	
	"github.com/fzzy/radix/redis"
)

type RedisConnection interface {
	Dial(protocol, uri string) (RedisClient,error)
}

type RedisReply interface {
	String() string
	Err() error
}

type RedisClient interface {
	Cmd(cmd string, args ...interface{}) RedisReply
}

type RadixRedisConnection struct {
	
}

type RadixRedisClient struct {
	client *redis.Client
}

type RadixRedisReply struct {
	reply *redis.Reply
}

func (c RadixRedisConnection) Dial(protocol, uri string) (RedisClient,error) {
	redisclient, err := redis.Dial(protocol, uri)
	return &RadixRedisClient{client : redisclient}, err
}

func (c *RadixRedisClient) Cmd(cmd string, args ...interface{}) RedisReply{
	re := c.client.Cmd(cmd, args)
	return &RadixRedisReply{reply : re}
}

func (c *RadixRedisReply) String() string{
	return c.reply.String()
}

func (c *RadixRedisReply) Err() error{
	return c.reply.Err
}
