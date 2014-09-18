package sentinel

import (
	//"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"time"
)

type SentinelHealthCheckerClient struct {
	sentinel        *types.Sentinel
	redisClient     *redis.Client
	sentinelManager *SentinelManager
	sleepInSeconds  int
}

func NewHealthCheckerClient(sentinel *types.Sentinel, manager *SentinelManager) (*SentinelHealthCheckerClient, error) {

	uri := sentinel.GetLocation()
	logger.Info.Print("HealthChecker : connecting to %s", uri)
	redisclient, err := redis.Dial("tcp", uri)

	if err != nil {
		logger.Info.Printf("HealthChecker : not connected to %s, %s", uri, err.Error())
		return nil, err
	}
	logger.Info.Printf("HealthChecker : connected to %s", uri)

	client := &SentinelHealthCheckerClient{redisClient: redisclient,
		sentinel:        sentinel,
		sentinelManager: manager,
		sleepInSeconds:  1}
	return client, nil
}

func (client *SentinelHealthCheckerClient) Start() {
	go client.loop()
}

func (client *SentinelHealthCheckerClient) loop() {

	for {

		r := client.redisClient.Cmd("PING")

		if r.Err != nil {

			client.sentinelManager.Notify(&SentinelLost{sentinel: client.sentinel})
		}

		pingResult := r.String()

		logger.Info.Printf("HealthChecker: %s says %s", client.sentinel.GetLocation(), pingResult)

		if pingResult != "PONG" {

			client.sentinelManager.Notify(&SentinelLost{sentinel: client.sentinel})

		} else {

			client.sentinelManager.Notify(&SentinelPing{sentinel: client.sentinel})
			//TODO : check for other sentinels
		}

		time.Sleep(time.Duration(client.sleepInSeconds) * time.Second)
	}
}

// func (client *SentinelHealthCheckerClient) findConnectedSentinels(clustername string) (bool, error) {
// 	r := client.redisclient.Cmd("SENTINEL", "SENTINELS", clustername)
// 	l := r.String()
// 	// TODO : Investigate why r.List() should return the correct datatype but doesn't
// 	// TODO : Parse into an array of arrays
// 	log.Printf("Sentinels : Sentinels : %s \n", l)
// 	return false, nil
// }
