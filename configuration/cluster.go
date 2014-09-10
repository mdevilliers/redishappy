package configuration

type Cluster struct {
	Name string
	Sentinels []Sentinel
	MasterPort int
	SlavePorts []int
}