package connect

import (
	"fmt"
	"net"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	defaultAutoReleaseInterval = time.Minute * 10
	defaultAutoReleaseMin      = 10
)

//Client ..
type Client struct {
	nt         string
	serverAddr *net.TCPAddr

	pool *pool

	intervalAutoRelease time.Duration
	minLimitAutoRelease uint32

	cacheMutex sync.Mutex
	cache      map[*origConn]bool

	isClosed bool
}

//NewClient ..
func NewClient(nt string, ip string, port int, max uint32) (*Client, error) {
	client := &Client{
		nt: nt,

		isClosed: false,
	}

	serverAddr, err := net.ResolveTCPAddr(nt, fmt.Sprintf("%s:%d", ip, port))
	if nil != err {
		return nil, err
	}

	client.serverAddr = serverAddr

	client.pool = newPool(max, client.dial)
	client.cache = make(map[*origConn]bool)

	client.intervalAutoRelease = defaultAutoReleaseInterval
	client.minLimitAutoRelease = defaultAutoReleaseMin

	return client, nil
}

//GetConn ..
func (client *Client) GetConn() (*Conn, error) {
	if client.isClosed {
		return newConnInErr(ErrClosed), ErrClosed
	}
	oc, err := client.getConn()
	if nil != err {
		return newConnInErr(err), err
	}
	c := newConn(oc)
	c.setReleaseFunc(client.releaseConn)
	return c, nil
}

//Close ..
func (client *Client) Close() bool {
	if client.isClosed {
		return false
	}
	client.isClosed = true
	client.closeAllConn()
	return true
}

func (client *Client) getConn() (*origConn, error) {
	return client.pool.getConn()
}

func (client *Client) dial() (*origConn, error) {
	tc, err := net.DialTCP(client.nt, nil, client.serverAddr)
	if nil != err {
		return nil, err
	}

	oc := newOrigConn(tc)

	client.cacheMutex.Lock()
	client.cache[oc] = true
	client.cacheMutex.Unlock()

	return oc, nil
}

func (client *Client) releaseConn(oc *origConn, needClose bool) {
	if client.isClosed || needClose {
		client.closeConn(oc)
		return
	}

	cached := client.pool.releaseConn(oc)
	if false == cached {
		client.closeConn(oc)
	}
}

func (client *Client) closeConn(oc *origConn) {
	client.cacheMutex.Lock()
	_, exist := client.cache[oc]
	if true == exist {
		delete(client.cache, oc)
	}
	client.cacheMutex.Unlock()

	if true == exist {
		oc.close()
		client.pool.closeOne()
	}
}

func (client *Client) closeAllConn() {
	client.cacheMutex.Lock()
	for oc := range client.cache {
		oc.close()
	}
	client.cacheMutex.Unlock()
}

//HoldingRelease ..
func (client *Client) HoldingRelease() {
	ticker := time.NewTicker(client.intervalAutoRelease)
	for !client.isClosed {
		<-ticker.C
		num := client.pool.count / 3
		if client.minLimitAutoRelease > client.pool.count-num {
			num = client.pool.count - client.minLimitAutoRelease
		}
		for i := uint32(0); i < num; i++ {
			oc, err := client.pool.getConnImmediately()
			if nil != err {
				break
			}
			if nil != oc {
				logs.Warn("Connection Client, HoldingRelease close connection, local: %v", oc.tc.LocalAddr().String())
				client.closeConn(oc)
			} else {
				break
			}
		}
	}
}

type pool struct {
	cache chan *origConn
	count uint32
	max   uint32

	newLock sync.Mutex
	fDial   func() (*origConn, error)
}

func newPool(max uint32, fd func() (*origConn, error)) *pool {
	p := &pool{
		cache: make(chan *origConn, max),
		count: 0,
		max:   max,

		fDial: fd,
	}
	return p
}

func (p *pool) getConn() (*origConn, error) {
	select {
	case oc, ok := <-p.cache:
		if false == ok {
			return nil, ErrClosed
		}
		return oc, nil
	default:
		if p.count < p.max {
			oc, err := p.newConn()
			if nil != err {
				return nil, err
			}
			return oc, nil
		}

		select {
		case oc, ok := <-p.cache:
			if false == ok {
				return nil, ErrClosed
			}
			return oc, nil
		case <-time.After(DefaultClientGetConnTimeout):
			return nil, ErrTimeout
		}
	}
}

func (p *pool) getConnImmediately() (*origConn, error) {
	select {
	case oc, ok := <-p.cache:
		if false == ok {
			return nil, ErrClosed
		}
		return oc, nil
	default:
		return nil, nil
	}
}

func (p *pool) newConn() (*origConn, error) {
	if nil == p.fDial {
		return nil, ErrInvalid
	}

	oc, err := p.fDial()
	if nil != err {
		return nil, err
	}

	p.newLock.Lock()
	defer p.newLock.Unlock()
	p.count++

	return oc, nil
}

func (p *pool) releaseConn(oc *origConn) bool {
	select {
	case p.cache <- oc:
		return true
	default:
		return false
	}
}

func (p *pool) closeOne() {
	p.newLock.Lock()
	defer p.newLock.Unlock()
	if 0 != p.count {
		p.count--
	}
}
