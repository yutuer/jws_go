package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"vcs.taiyouxi.net/platform/planx/client"
)

type P struct {
	X, Y, Z int
	Name    string
}

type Q struct {
	X, Y *int32
	Name string
}

func main() {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	dec := gob.NewDecoder(&network)
	// Encode (send) the value.
	err := enc.Encode(P{3, 4, 5, "Pythagoras"})
	if err != nil {
		log.Fatal("encode error:", err)
	}
	// Decode (receive) the value.
	var q Q
	err = dec.Decode(&q)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	fmt.Println(q)
	fmt.Printf("%q: {%d,%d}\n", q.Name, *q.X, *q.Y)

	fmt.Println("-=-=-=-=-=-")
	var network2 bytes.Buffer
	enc2 := gob.NewEncoder(&network2)
	dec2 := gob.NewDecoder(&network2)
	m := make(map[string]interface{})
	m["a"] = 1
	m["b"] = "a"
	err2 := enc2.Encode(&m)
	if err2 != nil {
		log.Fatal("encode error:", err2)
	}

	n := make(map[string]interface{})
	err2 = dec2.Decode(&n)
	if err2 != nil {
		log.Fatal("decode error:", err2)
	}
	fmt.Println(n)

	fmt.Println("-=-=-=-=-=-")
	var network3 bytes.Buffer
	enc3 := gob.NewEncoder(&network3)
	dec3 := gob.NewDecoder(&network3)
	err3 := enc3.Encode(client.PacketPing)
	if err3 != nil {
		log.Fatal("encode error:", err3)
	}
	var pkt client.Packet
	err3 = dec3.Decode(&pkt)
	if err3 != nil {
		log.Fatal("decode error:", err3)
	}
	fmt.Println(pkt)
}
