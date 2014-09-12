package configuration

type Cluster struct {
	Name       string
	MasterPort int
	SlavePorts []int
}
