package types

import (
	"testing"
)

func TestGetLocationMethod(t *testing.T) {

	sentinel := Sentinel{Host: "1.1.1.1", Port: 1234}

	location := sentinel.GetLocation()

	if location != "1.1.1.1:1234" {
		t.Errorf("Location string on wrong format : %s", location)
	}
}
