package client

import (
	"vcs.taiyouxi.net/jws/crossservice/util/discover"
)

//..
const (
	defaultProject = "comic"
	defaultVersion = "0.0.0.000"
	defaultService = "DefaultServiceName"
)

//Client ..
type Client struct {
	project string
	version string
	build   string
	service string
	index   string
	ip      string
	port    int
	extra   string

	regInfo discover.Service
}

//NewClient ..
func NewClient() *Client {
	c := &Client{
		project: defaultProject,
		version: defaultVersion,
		service: defaultService,
		index:   "",
		ip:      "",
		port:    0,
		extra:   "",
	}
	return c
}

func (c *Client) refreshServiceParam() discover.Service {
	c.regInfo = discover.Service{
		Project: c.project,
		Version: c.version,
		Build:   c.build,
		Service: c.service,
		Index:   c.index,
		IP:      c.ip,
		Port:    c.port,
		Extra:   c.extra,
	}
	return c.regInfo
}

//SetProject ..
func (c *Client) SetProject(p string) {
	c.project = p
}

//SetVersion ..
func (c *Client) SetVersion(v string) {
	c.version = v
}

//SetService ..
func (c *Client) SetService(s string) {
	c.service = s
}

//SetBuild ..
func (c *Client) SetBuild(build string) {
	c.build = build
}

//SetIndex ..
func (c *Client) SetIndex(i string) {
	c.index = i
}

//SetIP ..
func (c *Client) SetIP(ip string) {
	c.ip = ip
}

//SetPort ..
func (c *Client) SetPort(p int) {
	c.port = p
}

//SetExtra ..
func (c *Client) SetExtra(ex string) {
	c.extra = ex
}

//GetExtra ..
func (c *Client) GetExtra() string {
	return c.extra
}

//GetRegInfo ..
func (c *Client) GetRegInfo() discover.Service {
	c.refreshServiceParam()
	return c.regInfo
}

//Reg ..
func (c *Client) Reg() error {
	c.refreshServiceParam()
	return discover.RegService(&c.regInfo)
}

//UnReg ..
func (c *Client) UnReg() error {
	c.refreshServiceParam()
	return discover.UnRegService(&c.regInfo)
}
