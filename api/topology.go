package api

import (
	"net/http"

	"github.com/mdevilliers/redishappy/sentinel"
)

type TopologyApi struct {
	Manager *sentinel.SentinelManager
}

func (s *TopologyApi) Get(w http.ResponseWriter, r *http.Request) {
	t := s.Manager.GetCurrentTopology()
	responseJSON(w, t.Items())
}
