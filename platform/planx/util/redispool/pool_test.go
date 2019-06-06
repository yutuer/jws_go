package redispool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_Pool2(t *testing.T) {
	p := NewRedisPool(Config{
		HostAddr:       "127.0.0.1:6379",
		Capacity:       1,
		MaxCapacity:    100,
		IdleTimeout:    1 * time.Second,
		DBSelected:     10,
		DBPwd:          "",
		ConnectTimeout: 1 * time.Second,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
	}, "test", true)

	var waitter sync.WaitGroup
	for i := 0; i < 1000; i++ {
		waitter.Add(1)
		go func() {
			defer waitter.Done()
			c := p.Get()
			ac, _, _, _, _, _ := p.Stats()
			fmt.Printf("%d  ", ac)
			if c.IsNil() {
				fmt.Println("nil")
			}
			time.Sleep(time.Second)
			c.Close()
		}()
	}
	waitter.Wait()

	time.Sleep(time.Millisecond * 300)
	ac, _, _, _, _, _ := p.Stats()
	fmt.Println(ac)

	time.Sleep(time.Millisecond * 300)
	ac, _, _, _, _, _ = p.Stats()
	fmt.Println(ac)

	time.Sleep(time.Millisecond * 300)
	ac, _, _, _, _, _ = p.Stats()
	fmt.Println(ac)

	p.Close()
}
