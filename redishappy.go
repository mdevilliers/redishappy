package main

import (
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/kylelemons/go-gypsy/yaml"
	"io"
	"net/http"
	"net"
	"os"
	"text/template"
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
    // TODO

	//connect to the haproxy management socket
	buf := make([]byte, 512)

	socket := "/tmp/haproxy"
	sockettype := "unix" 
	conn, err := net.Dial(sockettype, socket)
	if err != nil {
    	panic(err)
	}   
	defer conn.Close()

	_, err = conn.Write([]byte("show info\n"))	
	if err != nil {
    	panic(err)
	}   

	n, err := conn.Read(buf)
	resp := make([]byte, n) 
    switch err {
	    case io.EOF:
	      resp = append(resp, buf[:n]...)
	    case nil:
	      resp = append(resp, buf[:n]...)
	    default:
	      panic(err)
    }

	fmt.Printf("haproxy at %s says '%s'\n", conn.RemoteAddr().String(), resp)

	// host a json endpoint
	fmt.Println("hosting json endpoint...")
	service := rpc.NewServer()
	service.RegisterCodec(json.NewCodec(), "application/json")
	service.RegisterService(new(HelloService), "")
	http.Handle("/rpc", service)
	http.ListenAndServe(":8085", nil)

}
