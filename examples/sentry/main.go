package main

import (
	//"errors"

	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	raven "github.com/getsentry/raven-go"
)

func main() {
	go test1()
	//go test3()

	time.Sleep(1000 * time.Millisecond)
}

func test1() {
	defer func() {
		if v := recover(); v != nil {
			log.Printf("test1 panic %v", v)
			trace := make([]byte, 2048)
			count := runtime.Stack(trace, true)
			tracej, _ := json.Marshal(string(trace[:count]))
			log.Printf("[GameCHANSerever] Stack of %d bytes: %s\n", count, tracej)
		}
	}()
	println("test1")
	panic(fmt.Errorf("test1 error"))
}

func panicCatch() {
	if v := recover(); v != nil {
		log.Printf("test2 panic %v", v)
		trace := make([]byte, 1024)
		count := runtime.Stack(trace, true)
		log.Printf("[GameCHANSerever] Stack of %d bytes: %s\n", count, trace)
	}
}

func test2() {
	//defer panicCatch()
	defer panicCaptureRun("test2")
	println("test2")
	panic(fmt.Errorf("test2 error"))
}

func test3() {
	test2()
}

func panicCaptureRun(prefix string) {

	// sentry DSN generated by Sentry server
	var sentryDSN string = "https://cce76e02e7124976967d90d71085263a:174cabdcfef346d6bafd9cb7fe70c901@app.getsentry.com/39313"

	client, err := raven.NewClient(sentryDSN, nil)
	if err != nil {
		log.Printf("panicCaptureRun ! %s", err.Error())
	}

	//var packet *raven.Packet
	//switch rval := recover().(type) {
	//case nil:
	//return
	//case error:
	//packet = raven.NewPacket(rval.Error(), raven.NewException(rval, raven.NewStacktrace(2, 3, nil)), nil)
	//default:
	//rvalStr := fmt.Sprint(rval)
	//packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)), nil)
	//}
	//client.Capture(packet, nil)

	// ... i.e. raisedErr is incoming error
	var raisedErr error
	raisedErr = recover().(error)
	//println("test" + raisedErr.Error())
	// create a stacktrace for our current context
	trace := raven.NewStacktrace(2, 3, nil)
	log.Printf("trace:%s", trace.Culprit())
	packet := raven.NewPacket(raisedErr.Error(), raven.NewException(raisedErr, trace))
	packet.Logger = "test"
	packet.ServerName = "testddd"
	packet.Level = "dddd"
	client.Capture(packet, nil)

	//eventID, ch := client.Capture(packet, nil)
	//if err = <-ch; err != nil {
	//logs.Critical("[Game]Sentry sent error failed, %s", err.Error())
	//}
	//message := fmt.Sprintf("Captured error with id %s: %q", eventID, raisedErr)
	//log.Println(message)
}
