package main

import (
	"fmt"
	"os"
	"text/template"
	"github.com/blackjack/syslog"

);

type Nonsense struct {
	Message string
}



func main(){
	fmt.Println("redis-happy started")

	// sys log test
	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslog.Syslog(syslog.LOG_INFO, "redis-happy started.")

	// load a configuration file

	// format a template
	data := Nonsense{"world"}
	tmpl, err := template.New("test").Parse("Hello {{.Message}}")
	if err != nil { panic(err) }
	err = tmpl.Execute(os.Stdout, data)
	if err != nil { panic(err) }

	// subscribe to redis

	// host a json endpoint

	
}