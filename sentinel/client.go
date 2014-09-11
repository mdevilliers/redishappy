package sentinel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"github.com/fzzy/radix/extra/pubsub"
	"github.com/fzzy/radix/redis"
)

type SentinelClient struct {
	subscriptionclient *pubsub.SubClient
	redisclient *redis.Client
}

type MasterSwitchedEvent struct {
	Name string
	OldMasterIp string
	OldMasterPort int
	NewMasterIp string 
	NewMasterPort int
}

func (event *MasterSwitchedEvent ) String() (string) {
	e, _ := json.Marshal(event)
	return string(e[:])
}

func NewClient(sentineladdr string) (*SentinelClient, error) { 
	
	redisclient, err := redis.Dial("tcp", sentineladdr)	
	if err != nil {
		return nil, err
	}

	redissubscriptionclient := pubsub.NewSubClient(redisclient)

	client := new(SentinelClient)
	client.redisclient = redisclient
	client.subscriptionclient = redissubscriptionclient

	return client, nil
}

func (client *SentinelClient) FindConnectedSentinels(clustername string) (bool, error){ 

	r := client.redisclient.Cmd("SENTINEL", "SENTINELS", clustername)
	l:= r.String()
	// TODO : Investigate why r.List() should return the correct datatype but doesn't
	// TODO : Parse into an array of arrays
	fmt.Printf( "Sentinels : Sentinels : %s \n", l)

	return false,nil
}

func (client *SentinelClient) StartMonitoring(switchmasterchannel chan MasterSwitchedEvent ) (error) {

	//TODO : fix radix client - doesn't support PSubscribe
	subr := client.subscriptionclient.Subscribe("+switch-master", "+slave-reconf-done ") 

	if subr.Err != nil{
		return subr.Err
	}

	go client.loopSubscription(switchmasterchannel)	

	return nil
}

func (sub *SentinelClient) loopSubscription( switchmasterchannel chan MasterSwitchedEvent){
	for {
		r := sub.subscriptionclient.Receive()
		if r.Timeout() {
			continue
		}
		if r.Err == nil {
			fmt.Printf( "Subscription Message : Channel : %s : %s\n", r.Channel, r.Message)

			if r.Channel == "+switch-master"{
				bits := strings.Split(r.Message, " ")

				oldmasterport, _ := strconv.Atoi(bits[2])
				newmasterport, _ := strconv.Atoi(bits[4])

				event := MasterSwitchedEvent{Name : bits[0], OldMasterIp: bits[1], OldMasterPort:oldmasterport, NewMasterIp:bits[3], NewMasterPort:newmasterport}
				switchmasterchannel <- event
			}
		}
	}
}