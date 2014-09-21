package sentinel

import (
	"errors"
	"github.com/mdevilliers/redishappy/services/logger"
	// "github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	// "reflect"
	"testing"
	// "time"
)

// see healthchecker_test for mocks used
func TestNewPubSubClientWillErrorWhenCanConnect(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{}}}

	_, err := NewPubSubClient(sentinel, redisConnection)

	if err != nil {
		t.Error("Client should not throw error if connected successfully!")
	}
}

func TestNewPubSubClientWillThrowErrorWhenCanNotConnect(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{Host: "DOESNOTEXIST", Port: 1234} // mock coded to not connect
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{}}}

	_, err := NewPubSubClient(sentinel, redisConnection)

	if err == nil {
		t.Error("Client should throw error if unable to connect!")
	}
}

// type TestPubSubClient struct {
// 	SubscribePubSubReply *TestRedisPubSubReply
// 	ReceivePubSubReply *TestRedisPubSubReply
// }

// type TestRedisPubSubReply struct {
// 	Error     error
// 	TimedOut bool
// 	ChannelListeningOn string
// 	MessageToReturn string
// }
func TestNewPubSubClientReturnsMasterSwitchEventToTheCorrectChannel(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	subscribeReply := &TestRedisPubSubReply{}
	subscriptionMessage := "name 1.1.1.1 1234 2.2.2.2 5678"
	receiveReply := &TestRedisPubSubReply{TimedOut: false, ChannelListeningOn: "+switch-master", MessageToReturn: subscriptionMessage}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{PubSubClient: &TestPubSubClient{SubscribePubSubReply: subscribeReply, ReceivePubSubReply: receiveReply}}}

	client, err := NewPubSubClient(sentinel, redisConnection)

	if err != nil {
		t.Error("Client should not throw error if connected successfully!")
	}

	switchmasterchannel := make(chan MasterSwitchedEvent)
	_ = client.StartMonitoringMasterEvents(switchmasterchannel)
	event := <-switchmasterchannel

	if event.Name != "name" {
		t.Error("Error recieving event")
	}
}

func TestNewPubSubClientReturnsErrorWhenConnectionDisappears(t *testing.T) {
	logger.InitLogging("../log")

	sentinel := types.Sentinel{}
	subscribeReply := &TestRedisPubSubReply{Error: errors.New("CANT'T CONNECT")}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{PubSubClient: &TestPubSubClient{SubscribePubSubReply: subscribeReply}}}

	client, err := NewPubSubClient(sentinel, redisConnection)

	if err != nil {
		t.Error("Client should not throw error if connected successfully!")
	}

	switchmasterchannel := make(chan MasterSwitchedEvent)
	err = client.StartMonitoringMasterEvents(switchmasterchannel)

	if err == nil {
		t.Error("PubSubClient should pass back error on failure")
	}
}

func TestParseMasterMessage(t *testing.T) {

	input := "name 1.1.1.1 1234 2.2.2.2 5678"
	event := parseSwitchMasterMessage(input)

	if event.Name != "name" {
		t.Error("Error parsing name")
	}
	if event.OldMasterIp != "1.1.1.1" {
		t.Error("Error parsing old master ip ")
	}
	if event.OldMasterPort != 1234 {
		t.Error("Error parsing old master port")
	}
	if event.NewMasterIp != "2.2.2.2" {
		t.Error("Error parsing new master ip")
	}
	if event.NewMasterPort != 5678 {
		t.Error("Error parsing new master port")
	}
}
