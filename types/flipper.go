package types

type MasterSwitchedEvent struct {
	Name          string
	OldMasterIp   string
	OldMasterPort int
	NewMasterIp   string
	NewMasterPort int
}

type MasterDetails struct {
	ExternalPort int
	Name         string
	Ip           string
	Port         int
}

type FlipperClient interface {
	InitialiseRunningState(state []MasterDetails)
	Orchestrate(switchEvent MasterSwitchedEvent)
}
