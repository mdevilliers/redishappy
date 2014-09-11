package main

import (
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net/http"
	"os"
	"text/template"
	"github.com/mdevilliers/redishappy/configuration"
	//"github.com/mdevilliers/redishappy/haproxy"
	"github.com/mdevilliers/redishappy/sentinel"
)

func main() {

	fmt.Println("redis-happy started")

	// sys log test
	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslog.Syslog(syslog.LOG_INFO, "redis-happy started.")

	configuration, err := configuration.LoadFromFile("config.json")

	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsed from config : %s\n", configuration.String())

	switchmasterchannel := make(chan sentinel.MasterSwitchedEvent)

	go loopSentinalEvents(switchmasterchannel)

	for _, configuredSentinel := range configuration.Sentinels {
		
		sentinelAddress := fmt.Sprintf("%s:%d",configuredSentinel.Host, configuredSentinel.Port)
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

func loopSentinalEvents( switchmasterchannel chan sentinel.MasterSwitchedEvent){

	for i := range switchmasterchannel{
		 		fmt.Printf("Master Switched : %s\n", i.String() )
	}
}

// func formatTempplateExample(){
// 	data := Nonsense{"world"}
// 	tmpl, err := template.New("test").Parse("Hello {{.Message}}\n")
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = tmpl.Execute(os.Stdout, data)
// 	if err != nil {
// 		panic(err)
// 	}
// }

//func contactHAProxyExample(){
	//connect to the haproxy management socket
	// client := haproxy.NewClient("/tmp/haproxy")    
	// response,_ := client.Rpc("show info\n")
	// fmt.Printf( "%s\n", response.Message)
	// response,_ = client.Rpc("show stat\n")
	// fmt.Printf( "%s\n", response.Message)
	// response,_ = client.Rpc("xxxx\n")
	// fmt.Printf( "%s\n", response.Message)
	// response,_ = client.Rpc("show acl\n")
	// fmt.Printf( "%s\n", response.Message)
//}

type Nonsense struct {
	Message string
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