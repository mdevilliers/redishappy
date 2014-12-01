package sentinel

import (
	"errors"
	"testing"

	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
)

type MockMessage struct {
	err         error
	messages    string
	channel     string
	messageType int
}

func (m *MockMessage) Err() error       { return m.err }
func (m *MockMessage) Message() string  { return m.messages }
func (m *MockMessage) Channel() string  { return m.channel }
func (m *MockMessage) MessageType() int { return m.messageType }

func TestMonitorSignalsAnError(t *testing.T) {

	switchmasterchannel := make(chan types.MasterSwitchedEvent)

	ok := dealWithSentinelMessage(&MockMessage{err: errors.New("Boom")}, switchmasterchannel)
	if !ok {
		t.Error("A boom error should have happened")
	}
}

func TestMonitorWillParseAndForwardOnAGoodMessage(t *testing.T) {

	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	validinput := "name 1.1.1.1 1234 2.2.2.2 5678"

	go func() {
		ok := dealWithSentinelMessage(&MockMessage{messageType: redis.Message, messages: validinput}, switchmasterchannel)
		if ok {
			t.Error("A valid message was passed")
		}
	}()

	event := <-switchmasterchannel

	if event.Name != "name" {
		t.Error("Error recieving event")
	}
}

func TestMonitorWillReturnFalseOnAnInvalidMessage(t *testing.T) {

	switchmasterchannel := make(chan types.MasterSwitchedEvent)
	invalidinput := "name 1.1.1.1 rubbish 2.2.2.2 5678"

	ok := dealWithSentinelMessage(&MockMessage{messageType: redis.Message, messages: invalidinput}, switchmasterchannel)
	if !ok {
		t.Error("An invalid message was passed")
	}
}

func TestParseMasterMessage(t *testing.T) {

	input := "name 1.1.1.1 1234 2.2.2.2 5678"
	event, err := parseSwitchMasterMessage(input)

	if err != nil {
		t.Error("Error parsing valid message")
	}

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

func TestParseInvalidMasterMessage(t *testing.T) {

	inputs := []string{
		"rubbish",
		"name 1.1.1.1 1234 2.2.2.2 rubbish",
		"name 1.1.1.1 rubbish 2.2.2.2 5566",
		"name rubbish 2.2.2.2 5566",
	}

	for _, input := range inputs {
		_, err := parseSwitchMasterMessage(input)

		if err == nil {
			t.Error("Error parsing invalid message")
		}
	}
}
