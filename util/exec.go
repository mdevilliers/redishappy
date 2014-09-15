package util

import (
	"os/exec"
)

func ExecuteCommand(cmd string) error {
	_, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return err
	}
	return nil
}