package types

import "errors"

type Consul struct {
	Datacenter string    `json:"datacenter"`
	Services   []Service `json:"services"`
}

type Service struct {
	Cluster string   `json:"cluster"`
	Node    string   `json:"node"`
	Tags    []string `json:"tags"`
}

func (c Consul) FindByClusterName(name string) (Service, error) {
	for _, service := range c.Services {
		if service.Cluster == name {
			return service, nil
		}
	}
	return Service{}, errors.New("Service not found")
}
