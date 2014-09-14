package configuration

import (
	"encoding/json"
)

type Sentinel struct {
	Host string
	Port int
}

func (s *Sentinel) String() string {
	e, _ := json.Marshal(s)
	return string(e[:])
}
