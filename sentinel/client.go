package sentinel

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/redis"
	"github.com/mdevilliers/redishappy/types"
	"strconv"
	"time"
)

type SentinelClient struct {
	sentinel        types.Sentinel
	redisClient     redis.RedisClient
	sentinelManager Manager
	sleepInSeconds  int
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
		sentinel:        sentinel,
		sentinelManager: manager,
		sleepInSeconds:  1}
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

func (client *SentinelClient) Start() {
	go client.healthcheckloop()
	// TODO : check for other sentinels
	//go client.sentineldiscoveryloop()
}

func (client *SentinelClient) healthcheckloop() {

	for {
		r := client.redisClient.Cmd("PING")

		if r.Err() != nil {
			client.sentinelManager.Notify(&SentinelLost{Sentinel: client.sentinel})
			break
		}

		pingResult := r.String()

		if pingResult != "PONG" {
			client.sentinelManager.Notify(&SentinelLost{Sentinel: client.sentinel})
			break
		} else {
			client.sentinelManager.Notify(&SentinelPing{Sentinel: client.sentinel})
		}
		time.Sleep(time.Duration(client.sleepInSeconds) * time.Second)
	}
}

// TODO : check for other sentinels
// func (client *SentinelClient) sentineldiscoveryloop() {
// 	for {
// 		client.findConnectedSentinels("secure")
// 		time.Sleep(time.Duration(client.sleepInSeconds) * time.Second)
// 	}
// }
// func (client *SentinelClient) findConnectedSentinels(clustername string) (bool, error) {
// 	r := client.redisClient.Cmd("SENTINEL", "SENTINELS", clustername)
// 	for _, e := range r.Elems {
// 		  t,_ := e.Hash()
// 		  logger.Info.Printf("Sentinels : xxx : %s : %s",t["ip"], t["port"])
// 	}
// 	return false, nil
// }
