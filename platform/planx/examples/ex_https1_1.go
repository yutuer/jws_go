package main

import (
	"log"
	"net/http"
)

// 两个server可以共用同一个ServeMux，因为里面有锁的实现

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {

	log.Printf("About to listen on 10443. Go to https://127.0.0.1:10443/")

	http.DefaultServeMux.HandleFunc("/", handler)

	go func() {
		server := &http.Server{}
		server.Handler = http.DefaultServeMux
		server.Addr = ":10443"
		err := server.ListenAndServeTLS("server.crt", "server.key")
		if err != nil {
			log.Fatal(err)
		}
	}()

	server := &http.Server{}
	server.Handler = http.DefaultServeMux
	server.Addr = ":8080"
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
