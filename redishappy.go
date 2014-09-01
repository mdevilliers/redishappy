package main

import (
	"fmt"
	"github.com/blackjack/syslog"
	"github.com/kylelemons/go-gypsy/yaml"
	"os"
	"text/template"
)

type Nonsense struct {
	Message string
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
	tmpl, err := template.New("test").Parse("Hello {{.Message}}")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}

	// subscribe to redis sentinal

	// host a json endpoint

}
