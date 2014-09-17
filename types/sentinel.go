package types

import (
	"fmt"
)

type Sentinel struct {
	Host string
	Port int
}

func (s *Sentinel) GetLocation() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
