package types

import (
	"net"
	"strconv"
)

type Sentinel struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (s *Sentinel) GetLocation() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}
