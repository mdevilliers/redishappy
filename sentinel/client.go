package sentinel

import (
	"errors"
	"fmt"
	"github.com/fzzy/radix/extra/pubsub"
	"github.com/fzzy/radix/redis"
)

type SentinelClient struct {
	subscriptionclient *pubsub.SubClient
}

func NewClient(sentineladdr string) (*SentinelClient, error) { 
	
	redisclient, err := redis.Dial("tcp", sentineladdr)	
	if err != nil {
		return nil, err
	}

	redissubscriptionclient := pubsub.NewSubClient(redisclient)

	client := new(SentinelClient)
	client.subscriptionclient = redissubscriptionclient

	return client, nil
}

func (client *SentinelClient) FindConnectedSentinels() (bool, error){ 
	return false,errors.New("Not Implemented!")
}

func (client *SentinelClient) StartMonitoring() (bool, error) {

	subr := client.subscriptionclient.Subscribe("+switch-master", "+slave-reconf-done ") //TODO : fix radix client - doesn't support PSubscribe

	if subr.Err != nil{
		return false, subr.Err
	}

	go client.loopSubscription()	

	return true, nil
}

func (sub *SentinelClient) loopSubscription(){
	for {
		r := sub.subscriptionclient.Receive()
		if r.Timeout() {
			continue
		}
		if r.Err == nil {
			fmt.Printf( "Subscription Message : Channel : %s : %s\n", r.Channel, r.Message)
		}
	}
}