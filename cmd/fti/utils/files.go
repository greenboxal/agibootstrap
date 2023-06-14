package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

// CreateDirIfNotExist creates a directory if it does not exist
func CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			return fmt.Errorf("failed to create directory: %v", errDir)
		}
	}
	return nil
}

// ReadFile reads a file and returns its content
func ReadFile(path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	return content, nil
}

// WriteFile writes content to a file
func WriteFile(path string, data []byte) error {
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}
