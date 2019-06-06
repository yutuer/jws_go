package main

import (
	"bytes"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type A struct {
	A1 int
	B1 int64
	C1 string
}

type B struct {
	A1 A
	B1 int64
	C1 string
}

type C struct {
	A
	B2 int64
	C2 string
}

func main() {
	a := A{-1, 100, "abc"}
	b := B{a, 101, "def"}
	c := C{a, 102, "ghi"}

	var buf bytes.Buffer
	fmt.Fprint(&buf, a)
	println(string(buf.Bytes()))

	fmt.Fprint(&buf, b)
	println(string(buf.Bytes()))

	db, err := redis.Dial("tcp", ":6379")
	if err != nil {
		// handle error
	}
	defer db.Close()
	//序列化到数据库， 借用了fmt.Fprint 所以b可以被序列化出去
	// https://github.com/garyburd/redigo/wiki/FAQ#does-redigo-provide-a-way-to-serialize-structs-to-redis
	db.Do("HMSET", redis.Args{}.Add("test:1").AddFlat(a)...)
	db.Do("HMSET", redis.Args{}.Add("test:2").AddFlat(b)...)
	db.Do("HMSET", redis.Args{}.Add("test:3").AddFlat(c)...)
	//Read B
	if v, err := redis.Values(db.Do("HGETALL", "test:2")); err != nil {
		fmt.Printf("redis values b error1,  %s\n", err.Error())
	} else {
		var bb B
		if err := redis.ScanStruct(v, &bb); err != nil {
			fmt.Printf("redis values b error2 %s\n", err.Error())
		}
	}

	//Read C
	if v, err := redis.Values(db.Do("HGETALL", "test:3")); err != nil {
		fmt.Printf("redis values b error1,  %s\n", err.Error())
	} else {
		var cc C
		if err := redis.ScanStruct(v, &cc); err != nil {
			fmt.Printf("redis values b error2 %s\n", err.Error())
		}
	}
}
