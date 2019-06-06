package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func ListenAndServeTLS(srv *http.Server, certFile, keyFile, clientca string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	pemByte, _ := ioutil.ReadFile(clientca)
	block, pemByte := pem.Decode(pemByte)
	cert, err1 := x509.ParseCertificate(block.Bytes)
	if err1 != nil {
		fmt.Println("ParseCertificate", err1.Error())
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	config.InsecureSkipVerify = true
	config.ClientAuth = tls.RequireAndVerifyClientCert
	config.ClientCAs = pool

	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
	return srv.Serve(tlsListener)
}

func main() {

	log.Printf("About to listen on 10443. Go to https://127.0.0.1:10443/")

	http.DefaultServeMux.HandleFunc("/", handler)

	server := &http.Server{}
	server.Handler = http.DefaultServeMux
	server.Addr = ":10443"
	err := ListenAndServeTLS(server, "server.crt", "server.key", "ca.crt")
	if err != nil {
		log.Fatal(err)
	}

}
