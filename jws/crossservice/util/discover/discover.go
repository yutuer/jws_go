package discover

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
)

//..
const (
	DefaultPathRoot = "Discover/registry"

	defaultTimeoutDial    = 5 * time.Second
	defaultTimeoutRequest = 3 * time.Second
)

var cfg client.Config
var discoverCfg Config

func init() {
	InitEtcdServerCfg(Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Root:      DefaultPathRoot,
	})
}

//Config ..
type Config struct {
	Endpoints []string
	Root      string
}

//InitEtcdServerCfg ..
func InitEtcdServerCfg(c Config) {
	cfg = client.Config{
		Endpoints: c.Endpoints,
		Transport: client.DefaultTransport,
	}
	discoverCfg = c
}

func callWithClient(f func(client.Client) error) error {
	cli, err := client.New(cfg)
	if nil != err {
		return err
	}
	return f(cli)
}

func makeServicePath(s *Service) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", discoverCfg.Root, s.Project, s.Version, s.Service, s.Index)
}

//MakeServicePathAsProject ..
func MakeServicePathAsProject(project string) string {
	return fmt.Sprintf("%s/%s", discoverCfg.Root, project)
}
