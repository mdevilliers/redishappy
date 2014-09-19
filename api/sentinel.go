package api

import (
	"encoding/json"
	"github.com/mdevilliers/redishappy/sentinel"
	"net/http"
)

type SentinelApi struct {
	Manager *sentinel.SentinelManager
}

func (s *SentinelApi) Get(w http.ResponseWriter, r *http.Request) {

	responseChannel := make(chan sentinel.SentinelTopology)
	s.Manager.GetState(sentinel.TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel
	responseJSON(w, topologyState)
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bites, _ := json.Marshal(data)
	w.Write(bites)
}
