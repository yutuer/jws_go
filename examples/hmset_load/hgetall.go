package main

import (
	"fmt"

	"flag"
	"sync"
	"sync/atomic"
	"time"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

var (
	_clients  int
	_requests int
	_host     string
	_port     int
	_password string
	_db       int
)

const c_key_format = "200:%d:99ca640f-1c2f-4e6a-b57c-2fd107139d55"

func init() {
	flag.IntVar(&_clients, "c", 50, "Number of parallel connections (default 50).")
	flag.IntVar(&_requests, "n", 100000, "Total number of requests (default 100000).")

	flag.StringVar(&_host, "h", "127.0.0.1", "Server hostname (default 127.0.0.1).")
	flag.IntVar(&_port, "p", 6379, "Server port (default 6379).")
	flag.StringVar(&_password, "a", "", "Password for Redis Auth.")
	flag.IntVar(&_db, "db", 0, "SELECT the specified db number (default 0).")
}

func main() {
	flag.Parse()
	total_start := time.Now().UnixNano()
	var nano1, nano2, nano3, nanotime2 int64
	var counter int64
	work := make(chan int, 256)
	var wg sync.WaitGroup
	for i := 0; i < _clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", _host, _port),
				redis.DialPassword(_password),
				redis.DialDatabase(_db),
			)
			if err != nil {
				panic(err)
			}
			for w := range work {
				key := fmt.Sprintf(c_key_format, w)
				//fmt.Println(len(rparams))
				beforew := time.Now().UnixNano()
				//_, err2 := conn.Do("HGETALL", key)
				//if err2 != nil {
				//	panic(err2)
				//}
				err3 := conn.Send("HGETALL", key)
				if err3 != nil {
					panic(err3)
				}

				do1 := time.Now().UnixNano() - beforew
				if err4 := conn.Flush(); err4 != nil {
					panic(err4)
				}

				do2 := time.Now().UnixNano() - beforew
				if _, err5 := conn.Receive(); err5 != nil {
					panic(err5)
				}

				do3 := time.Now().UnixNano() - beforew

				beforew2 := time.Now().UnixNano()
				//fmt.Println(id, ",", donew)
				donew2 := time.Now().UnixNano() - beforew2
				atomic.AddInt64(&nano1, int64(do1))
				atomic.AddInt64(&nano2, int64(do2))
				atomic.AddInt64(&nano3, int64(do3))
				atomic.AddInt64(&nanotime2, int64(donew2))
				atomic.AddInt64(&counter, 1)
			}
		}(i)
	}

	for i := 0; i < _requests; i++ {
		work <- i
	}
	close(work)
	wg.Wait()
	alldone := time.Now().UnixNano() - total_start
	fmt.Println("nano1:", _requests, " in ", nano1/int64(time.Second), "seconds, avg:", float64(nano1)/float64(int64(_requests)*int64(time.Millisecond)), " ms")
	fmt.Println("nano2:", _requests, " in ", nano2/int64(time.Second), "seconds, avg:", float64(nano2)/float64(int64(_requests)*int64(time.Millisecond)), " ms")
	fmt.Println("nano3:", _requests, " in ", nano3/int64(time.Second), "seconds, avg:", float64(nano3)/float64(int64(_requests)*int64(time.Millisecond)), " ms")
	//fmt.Println("Finish2:", _requests, " in ", nanotime2/int64(time.Second), "seconds, avg:", float64(nanotime2)/float64(int64(_requests)*int64(time.Millisecond)), " ms")
	fmt.Println("Finish3:", _requests, " in ", float64(alldone)/float64(time.Second), "seconds, avg:", float64(alldone)/float64(int64(_requests)*int64(time.Millisecond)), " ms")
}
