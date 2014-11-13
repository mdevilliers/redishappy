package main

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
)

type TemplateApi struct {
	ConfigurationManager *configuration.ConfigurationManager
}

func (s *TemplateApi) Get(w http.ResponseWriter, r *http.Request) {

	config := s.ConfigurationManager.GetCurrentConfiguration()
	contents, err := loadFile(config.HAProxy.TemplatePath)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	} else {
		io.WriteString(w, contents)
	}
}

type HAProxyApi struct {
	ConfigurationManager *configuration.ConfigurationManager
}

func (s *HAProxyApi) Get(w http.ResponseWriter, r *http.Request) {

	config := s.ConfigurationManager.GetCurrentConfiguration()
	contents, err := loadFile(config.HAProxy.OutputPath)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	} else {
		io.WriteString(w, contents)
	}
}

func loadFile(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		logger.Error.Printf("Error loading %s - %s", path, err.Error())
		return "", err
	}

	return string(contents), err
}
