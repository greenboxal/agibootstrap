package io

import (
	"io/ioutil"
)

// ReadFile reads a file and returns its contents as a string.
func ReadFile(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// WriteFile writes a string to a file.
func WriteFile(path string, contents string) error {
	return ioutil.WriteFile(path, []byte(contents), 0644)
}
