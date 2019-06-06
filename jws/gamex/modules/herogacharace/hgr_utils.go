package herogacharace

import (
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

func New2SecBackOff() *backoff.ExponentialBackOff {
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

type RedisDBSetting struct {
	AddrPort string
	Auth     string
	DB       int
}
