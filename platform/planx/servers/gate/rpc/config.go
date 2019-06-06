package rpc

type Config struct {
	//请指定内网IP，必须完整ip:port
	RPCListen       string `toml:"listen_addr"`
	RPCListenDocker string `toml:"listen_addr_docker"`
}
