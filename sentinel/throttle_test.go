package sentinel

import (
	"testing"

	"github.com/mdevilliers/redishappy/types"
)

func TestPassThrough(t *testing.T) {
	in := make(chan types.MasterSwitchedEvent)
	out := make(chan types.MasterSwitchedEvent)

	_ = NewThrottle(in, out)

	one := types.MasterSwitchedEvent{Name: "one", NewMasterIp: "2.2.2.2", NewMasterPort: 5678}

	in <- one
	event := <-out

	if event.Name != "one" {
		t.Error("First message not sent!")
	}
}

func TestDuplicateMessagesDropped(t *testing.T) {
	in := make(chan types.MasterSwitchedEvent, 5)
	out := make(chan types.MasterSwitchedEvent)

	_ = NewThrottle(in, out)

	one := types.MasterSwitchedEvent{NewMasterIp: "2.2.2.2", NewMasterPort: 5678}
	sameAsOne := types.MasterSwitchedEvent{NewMasterIp: "2.2.2.2", NewMasterPort: 5678}
	two := types.MasterSwitchedEvent{Name: "XXX", NewMasterIp: "9.9.9.9", NewMasterPort: 9999}

	in <- one
	in <- sameAsOne
	in <- two

	oneOut := <-out
	twoOut := <-out

	if oneOut.NewMasterIp == twoOut.NewMasterIp {
		t.Error("The second message should have been dropped")
	}
}
