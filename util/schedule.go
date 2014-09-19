package util

import (
	"time"
)

func Schedule(what func(), delay time.Duration) {
	go func() {
		time.Sleep(delay)
		what()
	}()
}
