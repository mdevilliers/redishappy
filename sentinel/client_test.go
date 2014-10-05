package sentinel

import (
	"errors"
	// "github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	// "github.com/mdevilliers/redishappy/types"
	"reflect"
	// "testing"
	// "time"
)

// MOCKS
type TestRedisConnection struct {
	RedisClient *TestRedisClient
}

type TestRedisClient struct {
	RedisReply   *TestRedisReply
	PubSubClient *TestPubSubClient
}

type TestRedisReply struct {
	Reply string
	Error error
}

type TestPubSubClient struct {
	SubscribePubSubReply *TestRedisPubSubReply
	ReceivePubSubReply   *TestRedisPubSubReply
}

type TestRedisPubSubReply struct {
	Error              error
	TimedOut           bool
	ChannelListeningOn string
	MessageToReturn    string
}

func (c TestRedisConnection) GetConnection(protocol, uri string) (redis.RedisClient, error) {

	//fail to connect
	if uri == "DOESNOTEXIST:1234" {
		return nil, errors.New("CannotConnect")
	}
	return c.RedisClient, nil
}

func (c *TestRedisClient) Cmd(cmd string, args ...interface{}) redis.RedisReply {
	return c.RedisReply
}

func (c *TestRedisClient) NewPubSubClient() redis.RedisPubSubClient {
	return c.PubSubClient
}

func (c *TestPubSubClient) Subscribe(channels ...interface{}) redis.RedisPubSubReply {
	return c.SubscribePubSubReply
}
func (c *TestPubSubClient) Receive() redis.RedisPubSubReply {
	return c.ReceivePubSubReply
}

func (c *TestRedisReply) String() string {
	return c.Reply
}

func (c *TestRedisReply) Err() error {
	return c.Error
}

func (c *TestRedisReply) List() ([]string, error) {
	return nil, nil
}

func (c *TestRedisReply) Hash() (map[string]string, error) {
	return nil, nil
}

func (c *TestRedisReply) Elems() []redis.RedisReply {
	return nil
}

func (c *TestRedisPubSubReply) Message() string {
	return c.MessageToReturn
}

func (c *TestRedisPubSubReply) Channel() string {
	return c.ChannelListeningOn
}

func (c *TestRedisPubSubReply) Timeout() bool {
	return c.TimedOut
}

func (c *TestRedisPubSubReply) Err() error {
	return c.Error
}

type TestManager struct {
	NotifyCalledWithSentinelPing  int
	NotifyCalledWithSentinelLost  int
	NotifyCalledWithSentinelAdded int
}

func (tm *TestManager) Notify(event SentinelEvent) {
	t := reflect.TypeOf(event).String()

	if t == "*sentinel.SentinelLost" {
		tm.NotifyCalledWithSentinelLost++
	}
	if t == "*sentinel.SentinelPing" {
		tm.NotifyCalledWithSentinelPing++
	}
	if t == "*sentinel.SentinelAdded" {
		tm.NotifyCalledWithSentinelAdded++
	}
}
func (tm *TestManager) GetState(request TopologyRequest) {

}

// func TestNewSentinelClientClientWillGetASuccessfulPing(t *testing.T) {
// 	logger.InitLogging("../log")

// 	sentinel := types.Sentinel{}
// 	sentinelManager := &TestManager{}
// 	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{Reply: "PONG"}}}

// 	client, _ := NewSentinelClient(sentinel, sentinelManager, redisConnection)
// 	client.Start()

// 	time.Sleep(time.Second)

// 	if sentinelManager.NotifyCalledWithSentinelPing != 1 {
// 		t.Error("Notify should have been called with a SentinelPing event!")
// 	}
// }

// func TestNewSentinelClientWillFailWhenPingUnsucessful(t *testing.T) {
// 	logger.InitLogging("../log")

// 	sentinel := types.Sentinel{}
// 	sentinelManager := &TestManager{}
// 	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{Reply: "ERROR"}}}

// 	client, _ := NewSentinelClient(sentinel, sentinelManager, redisConnection)
// 	client.Start()

// 	time.Sleep(time.Second)

// 	if sentinelManager.NotifyCalledWithSentinelLost != 1 {
// 		t.Error("Notify should have been called with a SentinelPing event!")
// 	}
// }

// func TestNewSentinelClientWillFailWhenErrorOnPing(t *testing.T) {
// 	logger.InitLogging("../log")

// 	sentinel := types.Sentinel{}
// 	sentinelManager := &TestManager{}
// 	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{Error: errors.New("BOOYAH!")}}}

// 	client, _ := NewSentinelClient(sentinel, sentinelManager, redisConnection)
// 	client.Start()

// 	time.Sleep(time.Second)

// 	if sentinelManager.NotifyCalledWithSentinelLost != 1 {
// 		t.Error("Notify should have been called with a SentinelPing event!")
// 	}
// }

// func TestNewSentinelClientWillWillSignalSentinelLostIfCanNotConnect(t *testing.T) {
// 	logger.InitLogging("../log")

// 	sentinel := types.Sentinel{Host: "DOESNOTEXIST", Port: 1234} // mock coded to not connect

// 	redisConnection := &TestRedisConnection{}

// 	_, _ = NewSentinelClient(sentinel, redisConnection)

// 	if sentinelManager.NotifyCalledWithSentinelLost != 1 {
// 		t.Error("Notify should have been called!")
// 	}
// }
