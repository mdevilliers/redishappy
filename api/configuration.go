package api

import (
	"net/http"

	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/util"
)

type ConfigurationApi struct {
	ConfigurationManager *configuration.ConfigurationManager
}

func (s *ConfigurationApi) Get(w http.ResponseWriter, r *http.Request) {
	config := s.ConfigurationManager.GetCurrentConfiguration()
	util.WriteResponseAsJSON(w, config)
}
