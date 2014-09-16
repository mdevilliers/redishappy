package main

import (
	"fmt"
	"io"
	"github.com/blackjack/syslog"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/natefinch/lumberjack"
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/services/flipper"
	"github.com/mdevilliers/redishappy/util"
	"log"
	"net/http"
	"os"
)

func main() {

	//TODO : configure from command line
	logPath := "log" //var/log/redis-happy")
	configFile := "config.json"

	initLogging(logPath)

	log.Print("redis-happy started")

	configuration, err := configuration.LoadFromFile(configFile)

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Parsed from config : %s\n", util.String(configuration))

	go initApiServer()

	flipper := flipper.New(configuration)

	switchmasterchannel := make(chan sentinel.MasterSwitchedEvent)

	go loopSentinelEvents(flipper, switchmasterchannel)

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
			log.Print(err)
		}

		sen.StartMonitoring(switchmasterchannel)
	}

}

func initApiServer(){

	log.Print("hosting json endpoint.")
	service := rpc.NewServer()
	service.RegisterCodec(json.NewCodec(), "application/json")
	service.RegisterService(new(HelloService), "")
	http.Handle("/rpc", service)
	http.ListenAndServe(":8085", nil)
}

func initLogging(logPath string) {
	if len(logPath) > 0 {
		
		syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
		syslogWriter := syslog.Writer{LogPriority: syslog.LOG_INFO}

		log.SetOutput(io.MultiWriter(&lumberjack.Logger{
			Dir:        logPath,
			NameFormat: "2006-01-02T15-04-05.000.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		}, os.Stdout, &syslogWriter ))
	}
}

func loopSentinelEvents(flipper * flipper.FlipperClient , switchmasterchannel chan sentinel.MasterSwitchedEvent) {

	for switchEvent := range switchmasterchannel {
		flipper.Orchestrate(switchEvent)
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
