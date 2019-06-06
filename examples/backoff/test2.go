package main

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
)

const (
	//三次, 总时间2s内
	DefaultInitialInterval     = 500 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 2
	DefaultMaxInterval         = 2 * time.Second
	DefaultMaxElapsedTime      = 2 * time.Second
)

func NewExponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		MaxElapsedTime:      DefaultMaxElapsedTime,
		Clock:               backoff.SystemClock,
	}
	if b.RandomizationFactor < 0 {
		b.RandomizationFactor = 0
	} else if b.RandomizationFactor > 1 {
		b.RandomizationFactor = 1
	}
	b.Reset()
	return b
}

func main() {
	b := NewExponentialBackOff()
	time.Sleep(b.NextBackOff())
	backoff.RetryNotify(func() error {
		fmt.Println("run")
		return fmt.Errorf("TEST")
	}, b, func(e error, d time.Duration) {
		fmt.Println(e, d)
	})
}
