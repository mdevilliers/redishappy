package util

import (
	"fmt"
	"testing"
)

func TestHashBytesAndEnsureCorrectAlgorithm(t *testing.T) {
	result := HashString("hello world")

	if result != "5eb63bbbe01eeed093cb22bb8f5acdc3" {
		t.Error("Hashing 'hello world' is not the expected value")
	}
}

func TestHashFile(t *testing.T) {
	result, err := HashFile("../readme.md")

	if err != nil {
		t.Error("Error hashing the readme - either the method doesn't work or the documentation is missing!")
	}

	fmt.Printf("%s", result)
}
