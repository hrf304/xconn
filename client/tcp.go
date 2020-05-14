package client

import (
	"github.com/google/uuid"
	"net"
	"xconn/common"
	"xconn/server"
	"xconn/tools"
)

type TcpConn struct{
	server.TcpConn
}

func newTcpConn(conn net.Conn, config *common.Config)*TcpConn {
	ci := &TcpConn{
	}
	ci.Conn = conn
	ci.Id = uuid.New().String()
	ci.RemoteAddress = conn.RemoteAddr().String()
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
	ci.DataSplitter = config.DataSplitter
	ci.PacketHandler = config.PacketHandler
	ci.Label = config.Label
	ci.IConn = ci

	return ci
}
