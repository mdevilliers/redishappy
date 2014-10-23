package sentinel

import (
	"errors"
	"testing"

	"github.com/mdevilliers/redishappy/types"
)

// see client_test for mocks used
func TestMonitorWillErrorWhenCanConnect(t *testing.T) {

	sentinel := types.Sentinel{}
	manager := &TestManager{}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{}}}

	_, err := NewMonitor(sentinel, manager, redisConnection)

	if err != nil {
		t.Error("Client should not throw error if connected successfully!")
	}
}

func TestMonitorWillThrowErrorWhenCanNotConnect(t *testing.T) {

	sentinel := types.Sentinel{Host: "DOESNOTEXIST", Port: 1234} // mock coded to not connect
	manager := &TestManager{}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{RedisReply: &TestRedisReply{}}}

	_, err := NewMonitor(sentinel, manager, redisConnection)

	if err == nil {
		t.Error("Client should throw error if unable to connect!")
	}
}

func TestMonitorReturnsMasterSwitchEventToTheCorrectChannel(t *testing.T) {

	sentinel := types.Sentinel{}
	manager := &TestManager{}
	subscribeReply := &TestRedisPubSubReply{}
	subscriptionMessage := "name 1.1.1.1 1234 2.2.2.2 5678"
	receiveReply := &TestRedisPubSubReply{TimedOut: false, ChannelListeningOn: "+switch-master", MessageToReturn: subscriptionMessage}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{PubSubClient: &TestPubSubClient{SubscribePubSubReply: subscribeReply, ReceivePubSubReply: receiveReply}}}

	client, err := NewMonitor(sentinel, manager, redisConnection)

	if err != nil {
		t.Error("Client should not throw error if connected successfully!")
	}

	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	_ = client.StartMonitoringMasterEvents(switchmasterchannel)
	event := <-switchmasterchannel

	if event.Name != "name" {
		t.Error("Error recieving event")
	}
}

func TestMonitorReturnsErrorWhenConnectionDisappears(t *testing.T) {

	sentinel := types.Sentinel{}
	manager := &TestManager{}
	subscribeReply := &TestRedisPubSubReply{Error: errors.New("CANT'T CONNECT")}
	redisConnection := &TestRedisConnection{RedisClient: &TestRedisClient{PubSubClient: &TestPubSubClient{SubscribePubSubReply: subscribeReply}}}

	client, err := NewMonitor(sentinel, manager, redisConnection)

	if err != nil {
		t.Error("Client should not throw error if connected successfully!")
	}

	switchmasterchannel := make(chan types.MasterSwitchedEvent)
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

func TestParseInstanceDetailsForIpAndPortMessage(t *testing.T) {

	input := "type name 1.1.1.1 1234 @ mastername 2.2.2.2 5678"
	host, port := parseInstanceDetailsForIpAndPortMessage(input)

	if host != "1.1.1.1" {
		t.Error("Error parsing ip")
	}
	if port != 1234 {
		t.Error("Error parsing port")
	}
}
