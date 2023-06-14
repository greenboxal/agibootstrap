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
	// Implement a function to read the configuration from a file
	// Read the config data from a file and return a Config struct
	// Example:
	//   file, err := os.Open("config.txt")
	//   if err != nil {
	//       fmt.Println("Error opening config file:", err)
	//       return Config{} // or handle the error as appropriate
	//   }
	//   defer file.Close()
	//   // Read the file content and parse it into a Config struct
	//   ...
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
