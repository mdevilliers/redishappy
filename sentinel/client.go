package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"strconv"
)

type SentinelClient struct {
	sentinel       types.Sentinel
	redisClient    redis.RedisClient
	manager        Manager
	sleepInSeconds int
}

func NewSentinelClient(sentinel types.Sentinel, manager Manager, redisConnection redis.RedisConnection) (*SentinelClient, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("SentinelClient : connecting to %s", uri)

	redisclient, err := redisConnection.GetConnection("tcp", uri)

	if err != nil {
		logger.Info.Printf("SentinelClient : not connected to %s, %s", uri, err.Error())
		manager.Notify(&SentinelLost{Sentinel: sentinel})
		return nil, err
	}

	logger.Info.Printf("SentinelClient : connected to %s", uri)

	client := &SentinelClient{redisClient: redisclient,
		sentinel:       sentinel,
		manager:        manager,
		sleepInSeconds: 1}
	return client, nil
}

func (m *SentinelClient) DiscoverMasterForCluster(clusterName string) (types.MasterDetails, error) {

	r := m.redisClient.Cmd("SENTINEL", "get-master-addr-by-name", clusterName)

	bits, err := r.List()

	if err == nil {
		port, _ := strconv.Atoi(bits[1])
		return types.MasterDetails{Name: clusterName, Ip: bits[0], Port: port}, nil
	}

	return types.MasterDetails{}, err
}

func (client *SentinelClient) FindConnectedSentinels(clustername string) {
	r := client.redisClient.Cmd("SENTINEL", "SENTINELS", clustername)
	for _, e := range r.Elems() {
		t, _ := e.Hash()
		logger.Info.Printf("Sentinel found : %s : %s", t["ip"], t["port"])
		port, _ := strconv.Atoi(t["port"])
		client.manager.Notify(&SentinelAdded{Sentinel: types.Sentinel{Host: t["ip"], Port: port}})
	}
}
