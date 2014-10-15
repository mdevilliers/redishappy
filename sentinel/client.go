package sentinel

import (
	"strconv"

	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
)

type SentinelClient struct {
	sentinel    types.Sentinel
	redisClient redis.RedisClient
}

func NewSentinelClient(sentinel types.Sentinel, redisConnection redis.RedisConnection) (*SentinelClient, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("SentinelClient : connecting to %s", uri)

	redisclient, err := redisConnection.GetConnection("tcp", uri)

	if err != nil {
		logger.Info.Printf("SentinelClient : not connected to %s, %s", uri, err.Error())
		return nil, err
	}

	logger.Info.Printf("SentinelClient : connected to %s", uri)

	client := &SentinelClient{redisClient: redisclient,
		sentinel: sentinel}
	return client, nil
}

func (m *SentinelClient) Close() {
	m.redisClient.Close()
}

func (m *SentinelClient) DiscoverMasterForCluster(clusterName string) (*types.MasterDetails, error) {

	r := m.redisClient.Cmd("SENTINEL", "get-master-addr-by-name", clusterName)

	bits, err := r.List()

	if err == nil {
		port, _ := strconv.Atoi(bits[1])
		return &types.MasterDetails{Name: clusterName, Ip: bits[0], Port: port}, nil
	}

	return &types.MasterDetails{}, err
}

func (client *SentinelClient) FindConnectedSentinels(clustername string) []types.Sentinel {

	r := client.redisClient.Cmd("SENTINEL", "SENTINELS", clustername)
	response := []types.Sentinel{}

	for _, e := range r.Elems() {

		t, _ := e.Hash()

		port, _ := strconv.Atoi(t["port"])
		response = append(response, types.Sentinel{Host: t["ip"], Port: port})
	}
	return response
}

func (client *SentinelClient) FindKnownClusters() []string {

	r := client.redisClient.Cmd("SENTINEL", "MASTERS")
	response := []string{}

	for _, e := range r.Elems() {

		t, _ := e.Hash()
		response = append(response, t["name"])
	}
	return response
}
