package main

import (
	"github.com/armon/consul-api"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type ConsulFlipperClient struct {
	consulClient *consulapi.Client
}

func NewConsulFlipperClient() *ConsulFlipperClient {

	client, err := consulapi.NewClient(consulapi.DefaultConfig())

	if err != nil {
		logger.Error.Printf("Error connecting to consul : %s", err.Error())
	}

	return &ConsulFlipperClient{consulClient: client}
}

func (c *ConsulFlipperClient) InitialiseRunningState(state *types.MasterDetailsCollection) {
	logger.Info.Printf("InitialiseRunningState called : %s", util.String(state.Items()))

	for _, md := range state.Items() {
		c.UpdateConsul(md.Name, md.Ip, md.Port)
	}
}

func (c *ConsulFlipperClient) Orchestrate(switchEvent types.MasterSwitchedEvent) {
	logger.Info.Printf("Orchestrate called : %s", util.String(switchEvent))

	c.UpdateConsul(switchEvent.Name, switchEvent.NewMasterIp, switchEvent.NewMasterPort)

}

func (c *ConsulFlipperClient) UpdateConsul(name string, ip string, port int) {

	//dig @127.0.0.1 -p 8600 testing.service.consul SRV
	catalog := c.consulClient.Catalog()

	service := &consulapi.AgentService{
		ID:      "redis-ha-example",
		Service: name,
		Tags:    []string{"redis", "master"},
		Port:    port,
	}

	reg := &consulapi.CatalogRegistration{
		Node:    "redis",
		Address: ip,
		Service: service,
	}

	_, err := catalog.Register(reg, nil)

	if err != nil {
		logger.Error.Printf("Error updating consul : %s", err.Error())
	}
}
