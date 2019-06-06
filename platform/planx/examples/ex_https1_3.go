package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {

	http.HandleFunc("/", handler)

	go func() {
		err := http.ListenAndServeTLS(":10443", "server.crt", "server.key", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}
