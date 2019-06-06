/*
AWS模块的Retryer机制配合AfterRetryHandler来实现。

cenk/backoff 模块和AWS模块AfterRetryHandler中的Retryer机制,基于次数的接口模式本身就不太一样。
因此cenk/backoff无法提前知道MaxRetries,所以无法直接使用AWS的默认Retryer套用.
因此使用cenk/backoff的时候需要让AWS默认Retry机制关闭才行。

因此需要实现自己的AfterRetryHandler

关闭AWS Retry时要注意
corehandlers/handlers:AfterRetryHandler中有关

		// when the expired token exception occurs the credentials
		// need to be expired locally so that the next request to
		// get credentials will trigger a credentials refresh.
		if r.IsErrorExpired() {
			r.Config.Credentials.Expire()
		}

的实现
*/
package awshelper

import (
	"time"

	"github.com/cenk/backoff"
)

/*
10 times, totally 2 mins.
tested in vcs.taiyouxi.net/examples/backoff
➜  backoff git:(develop) ✗ time go run test1.go
TEST 552.330144ms
TEST 1.440509088s
TEST 2.329120107s
TEST 3.750856749s
TEST 7.397099976s
TEST 18.989169166s
TEST 18.100384615s
TEST 39.391155284s
TEST 35.818171134s
go run test1.go  0.31s user 0.09s system 0% cpu 2:08.13 total
➜  backoff git:(develop) ✗ pwd

*/
const (
	DefaultInitialInterval     = 500 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 2
	DefaultMaxInterval         = 60 * time.Second
	DefaultMaxElapsedTime      = 2 * time.Minute
)

func NewExponentialBackOffSleepFirst() *backoff.ExponentialBackOff {
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
	time.Sleep(b.NextBackOff())
	return b
}
