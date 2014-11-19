package util

import (
	"encoding/json"
)

func String(v interface{}) string {
	e, _ := json.Marshal(v)
	return string(e[:])
}

func StringPrettify(v interface{}) string {
	e, _ := json.MarshalIndent(v, "", "  ")
	return string(e[:])
}
