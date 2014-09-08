package main

import (
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/kylelemons/go-gypsy/yaml"
	"net/http"
	"os"
	"text/template"
	"github.com/mdevilliers/redishappy/haproxy"
	"github.com/mdevilliers/redishappy/sentinel"
)

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

func main() {

	fmt.Println("redis-happy started")

	// sys log test
	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslog.Syslog(syslog.LOG_INFO, "redis-happy started.")

	// load a configuration file
	config, err := yaml.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	name, err := config.Get("name")

	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsed from config : %s\n", name)

	// format a template
	data := Nonsense{"world"}
	tmpl, err := template.New("test").Parse("Hello {{.Message}}\n")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}

	// subscribe to redis sentinel
    // and listen on the pubsub channel
 	sentinel, err := sentinel.NewClient("192.168.0.20:26379")
	if err != nil {
		panic(err)
	}

	sentinel.StartMonitoring()
	
	//connect to the haproxy management socket
	client := haproxy.NewClient("/tmp/haproxy")
    
	response,_ := client.Rpc("show info\n")
	fmt.Printf( "%s\n", response.Message)
	response,_ = client.Rpc("show stat\n")
	fmt.Printf( "%s\n", response.Message)
	response,_ = client.Rpc("xxxx\n")
	fmt.Printf( "%s\n", response.Message)
	response,_ = client.Rpc("show acl\n")
	fmt.Printf( "%s\n", response.Message)

	// host a json endpoint
	fmt.Println("hosting json endpoint...")
	service := rpc.NewServer()
	service.RegisterCodec(json.NewCodec(), "application/json")
	service.RegisterService(new(HelloService), "")
	http.Handle("/rpc", service)
	http.ListenAndServe(":8085", nil)

}

