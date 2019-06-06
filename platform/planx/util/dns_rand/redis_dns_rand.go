package dns_rand

import (
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	rd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

const (
	net_time_out = 100 * time.Millisecond
)

func GetAddrByDNS(addr string) string {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		logs.Error("[DNSRand] GetAddrByDNS conf err %s", err.Error())
		return addr
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		logs.Error("[DNSRand] GetAddrByDNS SplitHostPort err %s", err.Error())
		return addr
	}

	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	c := new(dns.Client)
	c.DialTimeout = net_time_out
	c.WriteTimeout = net_time_out
	c.ReadTimeout = net_time_out

	dns_addr := addresses(conf, c, host, port)
	if len(dns_addr) == 0 {
		logs.Error("[DNSRand] No address found for %s", addr)
		return addr
	}
	logs.Debug("[DNSRand] GetAddrByDNS success %s", dns_addr[0])
	return dns_addr[0]
}

func addresses(conf *dns.ClientConfig, c *dns.Client, host, port string) (ips []string) {
	m4 := new(dns.Msg)
	m4.SetQuestion(dns.Fqdn(host), dns.TypeA)
	servers := conf.Servers[:]

	for len(servers) > 0 {
		l := len(servers)
		i := rd.Int31n(int32(l))
		server := servers[i]
		if int(i) < l-1 {
			servers[i] = servers[l-1]
		}
		servers = servers[:l-1]

		d, _, err := c.Exchange(m4, net.JoinHostPort(server, conf.Port))
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Temporary() && e.Timeout() {
				logs.Warn("[DNSRand] addresses use server %s timeout", server)
			} else {
				logs.Error("[DNSRand] addresses use server %s err: %s", server, err.Error())
			}
			continue
		}

		if d.Rcode == dns.RcodeSuccess {
			for _, a := range d.Answer {
				switch a.(type) {
				case *dns.A:
					ips = append(ips,
						net.JoinHostPort(a.(*dns.A).A.String(), port))
				case *dns.AAAA:
					ips = append(ips,
						net.JoinHostPort(a.(*dns.AAAA).AAAA.String(), port))
				}
			}
			break
		} else {
			logs.Error("[DNSRand] Exchange fail")
		}
	}
	return ips
}
