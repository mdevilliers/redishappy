package main

import (
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/template"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"sync"
	"net/http"
	// "os"
)

func main() {

	fmt.Println("redis-happy started")

	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslog.Syslog(syslog.LOG_INFO, "redis-happy started.")

	configuration, err := configuration.LoadFromFile("config.json")

	if err != nil {
		panic(err)
	}

	// fmt.Printf("Parsed from config : %s\n", util.String(configuration))

	switchmasterchannel := make(chan sentinel.MasterSwitchedEvent)

	go loopSentinelEvents(switchmasterchannel, configuration)

	for _, configuredSentinel := range configuration.Sentinels {

		sentinelAddress := fmt.Sprintf("%s:%d", configuredSentinel.Host, configuredSentinel.Port)
		sen, err := sentinel.NewClient(sentinelAddress)

		// TODO : exploding is no good - needs to connect to at least one
		// sentinel. Also explore of any sentinel you have to find others
		// using   _ ,err = sen.FindConnectedSentinels("nameofcluster")
		// check against the list of clusters to validate you can find
		// an answer for all the clusters you are monitoring
		// Once this is all initilised then write an haproxy config that
		// validly documents the existing tompology
		if err != nil {
			panic(err)
		}

		sen.StartMonitoring(switchmasterchannel)
	}

	// host a json endpoint
	fmt.Println("hosting json endpoint...")
	service := rpc.NewServer()
	service.RegisterCodec(json.NewCodec(), "application/json")
	service.RegisterService(new(HelloService), "")
	http.Handle("/rpc", service)
	http.ListenAndServe(":8085", nil)

}

func loopSentinelEvents(switchmasterchannel chan sentinel.MasterSwitchedEvent, config *configuration.Configuration) {

	configuration := config
	lock := &sync.Mutex{}

	for masterSwitch := range switchmasterchannel {

		syslog.Syslogf(syslog.LOG_INFO, "redis cluster {%s} master failover detected from {%s}:{%d} to {%s}:{%d}.", masterSwitch.Name, masterSwitch.OldMasterIp, masterSwitch.OldMasterPort,masterSwitch.NewMasterIp, masterSwitch.NewMasterPort)
		
		fmt.Printf("Master Switched : %s\n",  util.String(masterSwitch))
		fmt.Printf("Current Configuration : %s\n",  util.String(configuration.Clusters))

		cluster, err := configuration.FindClusterByName(masterSwitch.Name)

		if err != nil {
			syslog.Syslogf(syslog.LOG_INFO, "redis cluster called %s not found in configuration.", masterSwitch.Name)
			return
		}
		
		fmt.Printf("Cluster Configuration : %s\n",  util.String(cluster))

		details := types.MasterDetails {
							ExternalPort: cluster.MasterPort,
							Name :masterSwitch.Name,
							Ip : masterSwitch.NewMasterIp,
							Port :masterSwitch.NewMasterPort}

		//render template
		// TODO : look into HAProxy supporting multiple config files....
		path := configuration.HAProxy.OutputPath 
		arr := []types.MasterDetails{details}
		renderedTemplate, err := template.RenderTemplate(path, &arr)

		if err != nil {
			syslog.Syslogf(syslog.LOG_INFO, "error rendering tempate at %s.", path)
			return
		}

		lock.Lock()
		defer lock.Unlock()

		//get hash of new config
		newFileHash := util.HashString(renderedTemplate)
		//check hash on existing config
		oldFileHash, err := util.HashFile(path)

		if err != nil {
			syslog.Syslogf(syslog.LOG_INFO, "error hashing existing haproxy config file at %s.", path)
			return
		}

		if newFileHash == oldFileHash {
			syslog.Syslogf(syslog.LOG_INFO, "existing config file up todate. New file hash :  %s == Old file hash %s", newFileHash, oldFileHash )
			return
		}

		//TODO : check we have permission to update file

		//update file
		err = template.WriteFile(path, renderedTemplate)
		//reload haproxy
		reloadCommand := configuration.HAProxy.ReloadCommand
		err = util.ExecuteCommand(reloadCommand)

		if err != nil {
			syslog.Syslogf(syslog.LOG_INFO, "error reloading haproxy with command %s.", reloadCommand)
			return
		}

		syslog.Syslog(syslog.LOG_INFO, "haproxy reload completed.")
	
	}
}

type HelloArgs struct {
	Who string
}

type HelloReply struct {
	Message string
}

type HelloService struct{}

func (h *HelloService) Say(r *http.Request, args *HelloArgs, reply *HelloReply) error {
	reply.Message = "Hello, " + args.Who + "!"
	return nil
}
