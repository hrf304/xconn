package client

import (
	"github.com/golang/glog"
	"github.com/google/uuid"
	"net"
	"xconn/common"
	"xconn/server"
	"xconn/tools"
)

type UdpConn struct{
	server.UdpConn
}

func newUdpConn(conn *net.UDPConn, addr *net.UDPAddr, config *common.Config)*UdpConn {
	localAddr := conn.LocalAddr().String()
	addrstr := addr.String()
	glog.Infoln(addrstr, localAddr)
	ci := &UdpConn{
	}
	ci.Conn = conn
	ci.UdpAddr = addr
	ci.Id = uuid.New().String()
	ci.RemoteAddress = addr.String()
	ci.LocalAddr = conn.LocalAddr().String()
	ci.Sender = tools.NewDataTransport(1, config.SendChanSize)
	ci.Receiver = tools.NewDataTransport(1, config.RecvChanSize)
	ci.Done = make(chan bool, 1)
	ci.TimeoutCheck = tools.NewTimeoutCheck(config.Interval, config.Timeout)
	if config.BufSize <= 0{
		config.BufSize = 1024
	}
	ci.RecvBufSize = config.BufSize
	ci.ConnCallback = config.ConnCallback
	ci.Label = config.Label
	ci.DataSplitter = config.DataSplitter
	ci.PacketHandler = config.PacketHandler
	ci.IConn = ci

	return ci
}
