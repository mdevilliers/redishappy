package api

import (
	"io"
	"net/http"
)

type PingApi struct {
}

func (p *PingApi) Get(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Pong")
}
