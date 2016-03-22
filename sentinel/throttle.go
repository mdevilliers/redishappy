package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
)

type Throttle struct {
	lastEvent *types.MasterSwitchedEvent
	in        chan types.MasterSwitchedEvent
	out       chan types.MasterSwitchedEvent
}

func NewThrottle(in chan types.MasterSwitchedEvent, out chan types.MasterSwitchedEvent) *Throttle {

	t := &Throttle{
		in:  in,
		out: out,
		lastEvent: &types.MasterSwitchedEvent{
			Name:          "unknown",
			OldMasterIp:   "unknown",
			OldMasterPort: 0,
			NewMasterIp:   "unknown",
			NewMasterPort: 0}}
	go t.loopEvents()
	return t
}

func (t *Throttle) loopEvents() {
	for {
		select {
		case event := <-t.in:
			if event.NewMasterIp != t.lastEvent.NewMasterIp || event.NewMasterPort != t.lastEvent.NewMasterPort {
				t.lastEvent.NewMasterIp = event.NewMasterIp
				t.lastEvent.NewMasterPort = event.NewMasterPort

				t.out <- event
			} else {
				logger.Trace.Printf("Deduped +switch-master event, as I've (probably) already done the same thing. %s : %s : %v", event.Name, event.NewMasterIp, event.NewMasterPort)
			}
		}
	}
}
