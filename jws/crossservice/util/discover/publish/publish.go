package publish

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
)

//..
const (
	DefaultPathRoot = "Discover/publish"

	defaultTimeoutRequest = 3 * time.Second
)

func init() {
	InitConfig(Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Root:      DefaultPathRoot,
	})
}

//Config ..
type Config struct {
	Endpoints []string
	Root      string
}

var cfg Config
var etcdCfg client.Config

//InitConfig ..
func InitConfig(c Config) {
	cfg = c
	etcdCfg = client.Config{
		Endpoints: c.Endpoints,
		Transport: client.DefaultTransport,
	}
}

func callWithClient(f func(client.Client) error) error {
	cli, err := client.New(etcdCfg)
	if nil != err {
		return err
	}
	return f(cli)
}

func makeParentPath(parent string) string {
	return fmt.Sprintf("%s/%s", cfg.Root, parent)
}
func makeElemPath(parent string, elem Elem) string {
	return fmt.Sprintf("%s/%s/%s", cfg.Root, parent, elem.Key)
}
