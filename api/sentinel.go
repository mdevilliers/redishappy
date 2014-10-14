package api

import (
	"encoding/json"
	"net/http"

	"github.com/mdevilliers/redishappy/sentinel"
)

type SentinelApi struct {
	Manager *sentinel.SentinelManager
}

func (s *SentinelApi) Get(w http.ResponseWriter, r *http.Request) {

	responseChannel := make(chan sentinel.SentinelTopology)
	s.Manager.GetState(sentinel.TopologyRequest{ReplyChannel: responseChannel})
	sentinelState := <-responseChannel
	responseJSON(w, sentinelState)
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bites, _ := json.Marshal(data)
	w.Write(bites)
}
