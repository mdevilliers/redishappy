package util

import (
	"encoding/json"
)

func String(v interface{}) string {
	e, _ := json.Marshal(v)
	return string(e[:])
}
