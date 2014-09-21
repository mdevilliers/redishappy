package types

type MasterSwitchedEvent struct {
	Name          string
	OldMasterIp   string
	OldMasterPort int
	NewMasterIp   string
	NewMasterPort int
}

type FlipperClient interface {
	Orchestrate(switchEvent MasterSwitchedEvent)
}
