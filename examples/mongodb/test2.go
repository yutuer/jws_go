package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	Name      string    `bson:",omitempty"`
	LastName  string    `bson:",omitempty"`
	Phone     string    `bson:",omitempty"`
	Other     string    `bson:",omitempty"`
	Timestamp time.Time `bson:",omitempty"`
}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("people2")

	index := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		log.Fatal(err)
	}

	index2 := mgo.Index{
		Key:        []string{"lastname"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index2)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Insert(&Person{
		Name: "Ale", Phone: "+55 53 8116 9639", Other: "1"},
		&Person{Name: "Cla", Phone: "+55 53 8402 8510", Other: "2"})
	if err != nil {
		log.Fatal(err)
	}

	change := bson.M{"$set": Person{Phone: "11111", Timestamp: time.Now()}}
	_, err = c.Upsert(Person{Name: "Ale"},
		change)

	if err != nil {
		log.Fatal("upsert", err)
	}

	result := Person{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Phone:", result.Phone)
	//fmt.Println("Live Servers:", session.LiveServers())
}
