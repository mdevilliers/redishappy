package types

type Cluster struct {
	Name       string
	MasterPort int
	SlavePorts []int
}
