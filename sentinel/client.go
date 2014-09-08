package sentinel

import (
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
	subr := redissubscriptionclient.Subscribe("+switch-master") //TODO : fix radix client - doesn't support PSubscribe

	if subr.Err != nil{
		return nil, err
	}

	client := new(SentinelClient)
	client.subscriptionclient = redissubscriptionclient

	return client, nil
}

func (client *SentinelClient) StartMonitoring() {
	go client.loopSubscription()	
}

func (sub *SentinelClient) loopSubscription(){
	for {
		r := sub.subscriptionclient.Receive()
		if r.Timeout() {
			continue
		}
		if r.Err == nil {
			fmt.Printf( "Subscription Message : Master changed  : %s\n", r.Message)
		}
	}
}