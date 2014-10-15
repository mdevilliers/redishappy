package types

import (
	"fmt"
)

type Sentinel struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (s *Sentinel) GetLocation() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
