package util

import (
	"os/exec"
)

func ExecuteCommand(cmd string) ([]byte, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}
