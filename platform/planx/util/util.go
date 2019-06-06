package util

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"

	"time"

	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MaxInt          = int(^uint(0) >> 1)
	MinInt          = -MaxInt - 1
	MaxUInt         = ^uint(0)
	ASyncCmdTimeOut = 5 * time.Second
)

func init() {

}

func Exit(code int) {
	logs.Close()
	os.Exit(code)
}

var netErrClosed = "use of closed network connection"

func NetIsClosed(err error) bool {
	if strings.HasSuffix(err.Error(), netErrClosed) {
		return true
	}
	return false
}

func RandByOffsetUInt32(main, offset uint32, r *rand.Rand) uint32 {
	if offset == 0 {
		return main
	}

	if offset < 0 {
		offset = -offset
	}

	rv := uint32(r.Int63n(int64(offset) * 2))
	if main+rv < offset {
		return 0
	}
	return main + rv - offset
}

func GetExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

const JsonPostTyp = "application/json; charset=utf-8"

func HttpPost(url, typ string, data []byte) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(url, typ, body)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	body_res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}

	return body_res, nil
}

func HttpPostWCode(url, typ string, data []byte) (int, []byte, error) {
	body := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(url, typ, body)
	if err != nil {
		return 404, []byte{}, err
	}

	defer resp.Body.Close()
	body_res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return 404, []byte{}, err
	}

	return resp.StatusCode, body_res, nil
}

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
