package util

import (
	"io/ioutil"
	"os"
)

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func WriteFile(outputFilePath string, content string) error {
	return ioutil.WriteFile(outputFilePath, []byte(content), 0666)
}
