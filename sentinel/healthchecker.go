package sentinel

import (
	//"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"time"
)

type SentinelHealthCheckerClient struct {
	sentinel        types.Sentinel
	redisClient     *redis.Client
	sentinelManager *SentinelManager
	sleepInSeconds  int
}

func NewHealthCheckerClient(sentinel types.Sentinel, manager *SentinelManager) (*SentinelHealthCheckerClient, error) {

	uri := sentinel.GetLocation()
	logger.Info.Printf("HealthChecker : connecting to %s", uri)
	redisclient, err := redis.Dial("tcp", uri)

	if err != nil {
		logger.Info.Printf("HealthChecker : not connected to %s, %s", uri, err.Error())
		scheduleNewHealthChecker(sentinel, manager)
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

			client.sentinelManager.Notify(&SentinelLost{Sentinel: &client.sentinel})
			scheduleNewHealthChecker(client.sentinel, client.sentinelManager)
			break
		}

		pingResult := r.String()

		// logger.Info.Printf("HealthChecker: %s says %s", client.sentinel.GetLocation(), pingResult)

		if pingResult != "PONG" {

			client.sentinelManager.Notify(&SentinelLost{Sentinel: &client.sentinel})
			scheduleNewHealthChecker(client.sentinel, client.sentinelManager)
			break
		} else {

			client.sentinelManager.Notify(&SentinelPing{Sentinel: &client.sentinel})
			//TODO : check for other sentinels
			//client.findConnectedSentinels("secure")
		}

		time.Sleep(time.Duration(client.sleepInSeconds) * time.Second)
	}
}

func scheduleNewHealthChecker(sentinel types.Sentinel, sentinelManager *SentinelManager) {
	logger.Info.Printf("HealthChecker : scheduling new healthChecker for , %s", util.String(sentinel))
	util.Schedule(func() { sentinelManager.NewSentinelMonitor(sentinel) }, time.Second*5)
}

// func (client *SentinelHealthCheckerClient) findConnectedSentinels(clustername string) (bool, error) {
// 	r := client.redisClient.Cmd("SENTINEL", "SENTINELS", clustername)
// 	for _, e := range r.Elems {
// 		  t,_ := e.Hash()
// 		  logger.Info.Printf("Sentinels : xxx : %s : %s",t["ip"], t["port"])
// 	}
// 	return false, nil
// }
