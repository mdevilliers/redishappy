package util

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
)

func HashFile(filepath string) (string, error) {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return HashString(string(contents)), nil
}

func HashString(text string) string {
	return HashBytes([]byte(text))
}

func HashBytes(bytes []byte) string {
	hasher := md5.New()
	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
