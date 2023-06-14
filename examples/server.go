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
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	var config Config
	json.Unmarshal(byteValue, &config)

	return config
}

type Config struct {
	Port         string `json:"port"`
	ContentType  string `json:"contentType"`
	DefaultRoute string `json:"defaultRoute"`
}
