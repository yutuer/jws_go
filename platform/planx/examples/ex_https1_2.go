package main

import (
	"log"
	"net/http"
)

//能在同一个server上进行双开的。看了go源码，目前1.3版本还是安全的。

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {
	// http.HandleFunc("/", handler)

	// go func() {
	//     err := http.ListenAndServeTLS(":10443", "server.crt", "server.key", nil)
	//     if err != nil {
	//         log.Fatal(err)
	//     }
	// }()

	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	//     log.Fatal(err)
	// }

	http.DefaultServeMux.HandleFunc("/", handler)
	server := &http.Server{}
	server.Handler = http.DefaultServeMux
	go func() {
		server.Addr = ":10443"
		err := server.ListenAndServeTLS("server.crt", "server.key")
		if err != nil {
			log.Fatal(err)
		}
	}()

	server.Addr = ":8080"
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
