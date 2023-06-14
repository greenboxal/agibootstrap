package main

type Server struct {
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})
	http.ListenAndServe(":8080", nil)
}
