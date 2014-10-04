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
	InitialiseRunningState(state *MasterDetailsCollection)
	Orchestrate(switchEvent MasterSwitchedEvent)
}

type MasterDetailsCollection struct {
	items map[string]*MasterDetails
}

func NewMasterDetailsCollection() MasterDetailsCollection {
	return MasterDetailsCollection{items: map[string]*MasterDetails{}}
}

func (m *MasterDetailsCollection) AddOrReplace(master *MasterDetails) {
	m.items[master.Name] = master
}

func (m *MasterDetailsCollection) Items() []*MasterDetails {

	arr := make([]*MasterDetails, 0, len(m.items))
	for _, value := range m.items {
		arr = append(arr, value)
	}
	return arr
}

type ByName []*MasterDetails

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
