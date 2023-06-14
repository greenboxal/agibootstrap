package rickroll

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", rickrollHandler)

	addr := ":8080" // Change this to the desired port number or address
	fmt.Printf("Server listening on %s...\n", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func rickrollHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", http.StatusSeeOther)
}
