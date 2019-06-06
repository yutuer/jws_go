package server

import (
	"fmt"
	"net"

	"vcs.taiyouxi.net/platform/planx/version"

	csCfg "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/helper"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/client"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Server ..
type Server struct {
	Gid        uint32
	ServerIP   string
	ServerPort int
	lAddr      *net.TCPAddr

	service  *service
	groupIDs []uint32

	ipFilter []string

	rs *client.Client
}

//NewServer ..
func NewServer(gid uint32, ip string, port int) *Server {
	s := &Server{
		Gid:        gid,
		ServerIP:   ip,
		ServerPort: port,
		groupIDs:   []uint32{},
	}
	return s
}

//AddGroupIDs ..
func (s *Server) AddGroupIDs(gl []uint32) {
	s.groupIDs = append(s.groupIDs, gl...)
}

//SetIPFilter ..
func (s *Server) SetIPFilter(fs []string) {
	s.ipFilter = fs
}

//Start ..
func (s *Server) Start() error {
	//启动监听
	service, err := newService(csCfg.GetIndex(), s.ServerIP, s.ServerPort, s.groupIDs)
	if nil != err {
		return fmt.Errorf("Server Start, Service build Failed, %v", err)
	}
	s.service = service
	s.service.setIPFilter(s.ipFilter)
	go s.service.run()
	s.lAddr = s.service.cs.LocalAddr()

	//服务注册
	param := &helper.ServiceParam{
		GroupIDs: s.groupIDs,
		IP:       s.lAddr.IP.String(),
		Port:     s.lAddr.Port,
		Gid:      s.Gid,
	}
	sParam, err := helper.MarshalServiceParam(param)
	if nil != err {
		return fmt.Errorf("Server Start, marshal Service Param Failed, %v", err)
	}

	handle := client.NewClient()
	handle.SetProject(helper.ProjectName)
	handle.SetService(helper.ServiceName)
	handle.SetVersion(version.Version)
	handle.SetBuild(fmt.Sprintf("[%v](%v)", version.BuildCounter, version.GitHash))
	handle.SetIP(s.lAddr.IP.String())
	handle.SetPort(s.lAddr.Port)
	handle.SetIndex(csCfg.GetIndex())
	handle.SetExtra(sParam)
	s.rs = handle
	if err := handle.Reg(); nil != err {
		s.service.close()
		return fmt.Errorf("Reg Service to Discover Failed, %v", err)
	}

	return nil
}

//Stop ..
func (s *Server) Stop() {
	//服务去注册
	logs.Warn("Cross Service, Stop Server, Cancel form Discover")
	if err := s.rs.UnReg(); nil != err {
		logs.Error(fmt.Sprintf("Cross Service, Stop Server, Cancel form Discover failed, %v", err))
	}
	//停止监听
	logs.Warn("Cross Service, Stop Server, Close Service")
	s.service.close()
}

//LocalAddr ..
func (s *Server) LocalAddr() *net.TCPAddr {
	return s.lAddr
}
