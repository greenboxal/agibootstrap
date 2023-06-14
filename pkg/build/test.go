package main

import (
	"fmt"
	"net/http"
)

type Server struct {
}

func main() {
	// TODO: Move logic to server struct
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})
	http.ListenAndServe(":8080", nil)
}
func orphanSnippet3() {
	server := &Server{}
	server.Start()

}
func orphanSnippet2() {
	http.HandleFunc("/", s.rootHandler)
	http.ListenAndServe(":8080", nil)

}
func orphanSnippet1() {
	fmt.Fprint(w, "Hello, world!")

}
func orphanSnippet0() {
	"fmt"
	"net/http"

}
