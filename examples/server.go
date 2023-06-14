package main

import (
	"fmt"
	"net/http"
)

func main() {
	config := readConfigFromFile()
	http.HandleFunc(config.DefaultRoute, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", config.ContentType)
		fmt.Fprint(w, "Hello, World!")
	})

	http.ListenAndServe(config.Port, nil)
}

func readConfigFromFile() Config {
	// TODO: Implement a function to read the configuration from a file
	return Config{
		Port:         ":8080",
		ContentType:  "text/html",
		DefaultRoute: "/",
	}
}

type Config struct {
	Port         string
	ContentType  string
	DefaultRoute string
}
