package api

import (
	"net/http"

	"github.com/mdevilliers/redishappy/configuration"
)

type ConfigurationApi struct {
	ConfigurationManager *configuration.ConfigurationManager
}

func (s *ConfigurationApi) Get(w http.ResponseWriter, r *http.Request) {
	config := s.ConfigurationManager.GetCurrentConfiguration()
	responseJSON(w, config)
}
