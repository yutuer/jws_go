package service

/*
import (
	"fmt"
	"net"
	"strings"

	rp "github.com/dropbox/godropbox/resource_pool"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ServiceOptions struct {
	rp.Options
}

type ServiceManager struct {
	rp.ResourcePool

	toClientChan  chan *client.Packet
	waitChanWrite *util.WaitGroupWrapper

	quit     chan struct{}
	Options  ServiceOptions
	services *util.LockMap
}

func parseResourceLocation(resourceLocation string) (
	name string,
	address string) {

	idx := strings.Index(resourceLocation, " ")
	if idx >= 0 {
		return resourceLocation[:idx], resourceLocation[idx+1:]
	}

	return "", resourceLocation
}

func NewServiceManager(options ServiceOptions) *ServiceManager {
	quit := make(chan struct{})
	pktChan := make(chan *client.Packet, 1024)
	wait := &util.WaitGroupWrapper{}
	if options.Open == nil {
		options.Open = func(res string) (interface{}, error) {
			_, addr := parseResourceLocation(res)
			con, err := net.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			pcagent := client.NewPacketConnAgent(
				res,
				*client.NewPacketConn(con),
			)
			go pcagent.Start(quit)
			wait.Wrap(
				func() {
					c := pcagent.GetReadingChan()
				A:
					for {
						select {
						case <-quit:
							break A
						case pkt, ok := <-c:
							if ok {
								pktChan <- pkt
							} else {
								logs.Trace("Gate.ServiceManager router routine return, due to upstream PacketConnAgent Chan closed. %s", pcagent.Name)
								return
							}
						}
					}
					logs.Trace("Gate.ServiceManager router routine exit. %s", pcagent.Name)
				})
			return pcagent, nil
		}
	}

	if options.Close == nil {
		options.Close = func(handle interface{}) error {
			logs.Trace("Gate.ServiceManager close handle")
			pcagent, ok := handle.(*client.PacketConnAgent)
			if !ok {
				logs.Trace("Gate.ServiceManager close handle2")
				return fmt.Errorf("Gate.ServiceManager, Pool Close error!")
			}
			pcagent.Stop()
			return nil
		}
	}

	return &ServiceManager{
		services:      util.NewLockMap(),
		Options:       options,
		ResourcePool:  rp.NewMultiResourcePool(options.Options, nil),
		toClientChan:  pktChan,
		waitChanWrite: wait,
		quit:          quit,
	}
}

func (s *ServiceManager) StartService(name, addr string) {
	res := name + " " + addr
	err := s.ResourcePool.Register(res)
	if err != nil {
		logs.Critical("Gate.ServiceManager Start a service which has already registered!")
		return
	}

	s.services.Set(name, res)
}

func (s *ServiceManager) Stop() {
	s.waitChanWrite.Wait()
	close(s.toClientChan)
}

func (s *ServiceManager) GetPacketChan() <-chan *client.Packet {
	return s.toClientChan
}

func (s *ServiceManager) SendPacket(name string, pkt *client.Packet) error {
	res, ok := s.services.Get(name)
	if ok {
		// 0 获取资源
		MangeHandle, err := s.ResourcePool.Get(res.(string))
		if err != nil {
			logs.Error("Gate.ServiceManger %s SendPacket error 1, %s", name, err.Error())
			return err
		}

		// 1 使用资源
		handle, err := MangeHandle.Handle()
		if err != nil {
			logs.Error("Gate.ServiceManger %s SendPacket error 2, %s", name, err.Error())
			MangeHandle.Release()
			return err
		}

		if agent, ok := handle.(*client.PacketConnAgent); ok {
			//这里无法使用Send，因为无法得知链接级别错误，然后回收掉资源
			if err := agent.SendPacket(pkt); err != nil {
				logs.Error("Gate.ServiceManger %s SendPacket error 3, %s", name, err.Error())
				MangeHandle.Discard()
				return err
			}
		}

		// 2 释放资源
		s.ResourcePool.Release(MangeHandle)
		return nil
	}
	return fmt.Errorf("Gate.ServiceManager cant find service name %s", name)

}

*/
