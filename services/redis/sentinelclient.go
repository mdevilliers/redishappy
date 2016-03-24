package redis

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
)

type SentinelClient struct {
	sentinel    types.Sentinel
	redisClient RedisClient
}

func NewSentinelClient(sentinel types.Sentinel, redisConnection RedisConnection, tcpKeepAlive int) (*SentinelClient, error) {

	uri := sentinel.GetLocation()

	redisclient, err := redisConnection.GetConnection("tcp", uri, tcpKeepAlive)

	if err != nil {
		logger.Info.Printf("SentinelClient : not connected to %s, %s", uri, err.Error())
		return nil, err
	}

	client := &SentinelClient{
		redisClient: redisclient,
		sentinel:    sentinel,
	}

	return client, nil
}

func (m *SentinelClient) Close() {
	m.redisClient.ClosePool()
}

func (m *SentinelClient) Ping() error {
	return m.redisClient.Ping()
}

func (m *SentinelClient) DiscoverMasterForCluster(clusterName string) (*types.MasterDetails, error) {

	r, err := m.redisClient.SentinelGetMaster(clusterName)
	if err != nil {
		return nil, err
	}
	return &types.MasterDetails{Name: clusterName, Ip: r.Host, Port: r.Port}, nil
}

func (client *SentinelClient) FindConnectedSentinels(clustername string) ([]types.Sentinel, error) {

	r, err := client.redisClient.SentinelSentinels(clustername)
	response := []types.Sentinel{}

	if err != nil {
		return response, err
	}

	for _, s := range r {
		response = append(response, types.Sentinel{Host: s.IP, Port: s.Port})
	}
	return response, nil
}

func (client *SentinelClient) FindKnownClusters() ([]string, error) {

	r, err := client.redisClient.SentinelMasters()
	response := []string{}

	if err != nil {
		return response, err
	}

	for _, c := range r {
		response = append(response, c.Name)
	}
	return response, nil
}
