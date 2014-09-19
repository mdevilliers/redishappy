package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"testing"
	"time"
)

type TestRedisConnection struct {
	// RedisClient *TestRedisClient
}

type TestRedisClient struct{}

type TestRedisReply struct{}

func (c TestRedisConnection) Dial(protocol, uri string) (redis.RedisClient, error) {
	return &TestRedisClient{}, nil
}

func (c *TestRedisClient) Cmd(cmd string, args ...interface{}) redis.RedisReply {
	return &TestRedisReply{}
}

func (c *TestRedisReply) String() string {
	return "PIG"
}

func (c *TestRedisReply) Err() error {
	return nil
}

type TestManager struct {
	NotifyCalled                   int
	ScheduleNewHealthCheckerCalled int
}

func (tm *TestManager) Notify(event SentinelEvent) {
	tm.NotifyCalled++
}
func (*TestManager) GetState(request TopologyRequest) {

}
func (*TestManager) NewSentinelMonitor(types.Sentinel) (*SentinelHealthCheckerClient, error) {
	return nil, nil
}
func (tm *TestManager) ScheduleNewHealthChecker(sentinel types.Sentinel) {
	tm.ScheduleNewHealthCheckerCalled++
}

func TestNewClientWillFailWhenPingUnsucessful(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	sentinelManager := &TestManager{}
	redisConnection := &TestRedisConnection{}

	client, _ := NewHealthCheckerClient(sentinel, sentinelManager, redisConnection)
	client.Start()

	time.Sleep(time.Second)

	if sentinelManager.NotifyCalled != 1 {
		t.Error("Notify should have been called!")
	}
}
