package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	config := readConfigFromFile()
	db := setupDatabase()

	http.HandleFunc(config.DefaultRoute, func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := getValue(db, key)
		if value != "" {
			fmt.Fprintf(w, "Value for key %q: %s", key, value)
		} else {
			value := r.URL.Query().Get("value")
			setValue(db, key, value)
			fmt.Fprintf(w, "Successfully set value %q for key %q", value, key)
		}
	})

	http.ListenAndServe(config.Port, nil)
}

func setValue(db *gorm.DB, key string, value string) {
	db.Create(&KeyValue{Key: key, Value: value})
}

type Config struct {
	Port         string `json:"port"`
	ContentType  string `json:"contentType"`
	DefaultRoute string `json:"defaultRoute"`
}

func setupDatabase() *gorm.DB {
	db, err := gorm.Open("sqlite3", "store.db")
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&KeyValue{})

	return db
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
