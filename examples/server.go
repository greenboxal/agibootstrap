package main

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "Hello, World!")
	})

	http.ListenAndServe(":8080", nil)
}
