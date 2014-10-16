package util

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {

	file, err := ioutil.TempFile("", "")
	defer file.Close()
	defer os.Remove(file.Name())

	if err != nil {
		t.Error("Error creating temp file")
	}
	ok := FileExists(file.Name())

	if !ok {
		t.Error("File does exist - what is going on?")
	}
}

func TestFileExistsFails(t *testing.T) {

	ok := FileExists("does-not-exist")

	if ok {
		t.Error("File does not exist - what is going on?")
	}
}
