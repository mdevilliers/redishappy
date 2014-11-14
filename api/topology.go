package api

import (
	"net/http"

	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/util"
)

type TopologyApi struct {
	Manager *sentinel.SentinelManager
}

func (s *TopologyApi) Get(w http.ResponseWriter, r *http.Request) {
	t := s.Manager.GetCurrentTopology()
	util.WriteResponseAsJSON(w, t.Items())
}
