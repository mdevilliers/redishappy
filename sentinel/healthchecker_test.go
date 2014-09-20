package sentinel

import (
	"errors"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"reflect"
	"testing"
	"time"
)


func TestNewClientWillGetASuccessfulPing(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	sentinelManager := &TestManager{}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{Reply: "PONG"}}}

	client, _ := NewHealthCheckerClient(sentinel, sentinelManager, redisConnection)
	client.Start()

	time.Sleep(time.Second)

	if sentinelManager.NotifyCalledWithSentinelPing != 1 {
		t.Error("Notify should have been called with a SentinelPing event!")
	}
}

func TestNewClientWillFailWhenPingUnsucessful(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	sentinelManager := &TestManager{}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{Reply: "ERROR"}}}

	client, _ := NewHealthCheckerClient(sentinel, sentinelManager, redisConnection)
	client.Start()

	time.Sleep(time.Second)

	if sentinelManager.NotifyCalledWithSentinelLost != 1 {
		t.Error("Notify should have been called with a SentinelPing event!")
	}
}

func TestNewClientWillFailWhenErrorOnPing(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	sentinelManager := &TestManager{}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{Error: errors.New("BOOYAH!")}}}

	client, _ := NewHealthCheckerClient(sentinel, sentinelManager, redisConnection)
	client.Start()

	time.Sleep(time.Second)

	if sentinelManager.NotifyCalledWithSentinelLost != 1 {
		t.Error("Notify should have been called with a SentinelPing event!")
	}
}

func TestNewClientWillWillSignalSentinelLostIfCanNotConnect(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{Host: "DOESNOTEXIST", Port: 1234} // mock coded to not connect
	sentinelManager := &TestManager{}
	redisConnection := &TestRedisConnection{}

	_, _ = NewHealthCheckerClient(sentinel, sentinelManager, redisConnection)

	if sentinelManager.NotifyCalledWithSentinelLost != 1 {
		t.Error("Notify should have been called!")
	}
}

// MOCKS
type TestRedisConnection struct {
	RedisClient *TestRedisClient
}

type TestRedisClient struct {
	RedisReply *TestRedisReply
}

type TestRedisReply struct {
	Reply string
	Error error
}

func (c TestRedisConnection) Dial(protocol, uri string) (redis.RedisClient, error) {

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
	return nil
}

func (c *TestRedisReply) String() string {
	return c.Reply
}

func (c *TestRedisReply) Err() error {
	return c.Error
}

type TestManager struct {
	NotifyCalledWithSentinelPing   int
	NotifyCalledWithSentinelLost   int
	NotifyCalledWithSentinelAdded  int
	ScheduleNewHealthCheckerCalled int
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
func (*TestManager) GetState(request TopologyRequest) {

}
func (*TestManager) NewSentinelMonitor(types.Sentinel) (*SentinelHealthCheckerClient, error) {
	return nil, nil
}
func (tm *TestManager) ScheduleNewHealthChecker(sentinel types.Sentinel) {
	tm.ScheduleNewHealthCheckerCalled++
}