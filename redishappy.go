package main

import (
	"fmt"
	"github.com/blackjack/syslog"
);

func main(){
	fmt.Println("redis-happy started")

	// sys log test
	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslog.Syslog(syslog.LOG_INFO, "redis-happy started.")

	// load a configuration file

	// format a template

	// subscribe to redis

	// host a json endpoint

	
}