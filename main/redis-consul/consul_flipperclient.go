package main

import (
	"fmt"

	"github.com/armon/consul-api"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type ConsulFlipperClient struct {
	consulClient         *consulapi.Client
	configurationManager *configuration.ConfigurationManager
}

func NewConsulFlipperClient(cm *configuration.ConfigurationManager) *ConsulFlipperClient {

	configuration := cm.GetCurrentConfiguration()
	connectionDetails := consulapi.DefaultConfig()

	if configuration.Consul.Address != "" {
		connectionDetails.Address = configuration.Consul.Address
	}

	client, err := consulapi.NewClient(connectionDetails)

	if err != nil {
		logger.Error.Panicf("Error connecting to consul : %s", err.Error())
	}

	return &ConsulFlipperClient{consulClient: client, configurationManager: cm}
}

func (c *ConsulFlipperClient) InitialiseRunningState(state *types.MasterDetailsCollection) {

	logger.Info.Printf("InitialiseRunningState called : %s", util.String(state.Items()))
	for _, md := range state.Items() {
		c.UpdateConsul(md.Name, md.Ip, md.Port)
	}
}

func (c *ConsulFlipperClient) Orchestrate(switchEvent types.MasterSwitchedEvent) {

	logger.NoteWorthy.Printf("Redis master changed : %s", util.String(switchEvent))
	c.UpdateConsul(switchEvent.Name, switchEvent.NewMasterIp, switchEvent.NewMasterPort)

}

func (c *ConsulFlipperClient) UpdateConsul(name string, ip string, port int) {

	configuration := c.configurationManager.GetCurrentConfiguration()
	consulDetails := configuration.Consul

	service, err := consulDetails.FindByClusterName(name)

	if err != nil {
		logger.Error.Printf("Error locating service %s, %s", name, err.Error())
	}

	//dig @127.0.0.1 -p 8600 testing.service.consul SRV
	catalog := c.consulClient.Catalog()

	consulService := &consulapi.AgentService{
		ID:      fmt.Sprintf("redishappy-consul-%s", name),
		Service: name,
		Tags:    service.Tags,
		Port:    port,
	}

	reg := &consulapi.CatalogRegistration{
		Datacenter: service.Datacenter,
		Node:       service.Node,
		Address:    ip,
		Service:    consulService,
	}

	_, err = catalog.Register(reg, nil)

	if err != nil {
		logger.Error.Printf("Error updating consul : %s", err.Error())
	}
}
