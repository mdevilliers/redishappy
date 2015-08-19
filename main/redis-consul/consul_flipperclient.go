package main

import (
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/mdevilliers/golang-bestiary/pkg/retry"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"

	"golang.org/x/net/context"
)

type ConsulFlipperClient struct {
	sync.RWMutex
	consulClient         *api.Client
	clientConnected      bool
	configurationManager *configuration.ConfigurationManager
}

func NewConsulFlipperClient(cm *configuration.ConfigurationManager) *ConsulFlipperClient {

	return &ConsulFlipperClient{clientConnected: false, configurationManager: cm}
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

func (c *ConsulFlipperClient) getConnectedClient() (*api.Client, error) {

	c.Lock()
	defer c.Unlock()

	if c.clientConnected {
		return c.consulClient, nil
	}

	var wg sync.WaitGroup
	var errorToReurn error = nil

	retryAble := func() error {

		configuration := c.configurationManager.GetCurrentConfiguration()
		connectionDetails := api.DefaultConfig()

		if configuration.Consul.Address != "" {
			connectionDetails.Address = configuration.Consul.Address
		}

		client, err := api.NewClient(connectionDetails)

		if err != nil {

			return fmt.Errorf("Error connecting to consul : %s", err.Error())
		}

		c.consulClient = client
		c.clientConnected = true
		return nil
	}

	wg.Add(1)

	var notifee = func(err error) {

		errorToReurn = err
		wg.Done()
	}

	ret := retry.NewRetry(retryAble, 5, 1000) // total max retry of 31s
	ret.Execute(context.Background(), notifee)
	wg.Wait()

	return c.consulClient, errorToReurn

}

func (c *ConsulFlipperClient) UpdateConsul(name string, ip string, port int) {

	/// TODO : lock this access
	/// use the canceable context to ensure the latest update is the only one applied
	retryAble := func() error {

		configuration := c.configurationManager.GetCurrentConfiguration()
		client, err := c.getConnectedClient()

		if err != nil {

			return fmt.Errorf("Error connecting to consul : %s", err.Error())
		}

		consulDetails := configuration.Consul

		service, err := consulDetails.FindByClusterName(name)

		if err != nil {
			return fmt.Errorf("Error locating service %s, %s", name, err.Error())
		}

		//dig @127.0.0.1 -p 8600 testing.service.consul SRV
		catalog := client.Catalog()

		consulService := &api.AgentService{
			ID:      fmt.Sprintf("redishappy-consul-%s", name),
			Service: name,
			Tags:    service.Tags,
			Port:    port,
		}

		reg := &api.CatalogRegistration{
			Datacenter: service.Datacenter,
			Node:       service.Node,
			Address:    ip,
			Service:    consulService,
		}

		_, err = catalog.Register(reg, nil)

		if err != nil {
			return fmt.Errorf("Error updating consul : %s", err.Error())
		}
		return nil
	}

	var notifee = func(err error) {

	}

	ret := retry.NewRetry(retryAble, 5, 1000)
	ret.Execute(context.Background(), notifee)
}
