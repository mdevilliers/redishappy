package util

import (
	"encoding/json"
	"net/http"
)

func WriteResponseAsJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bites, _ := json.Marshal(data)
	w.Write(bites)
}
