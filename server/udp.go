package server

import (
	"github.com/golang/glog"
	"github.com/google/uuid"
	"net"
	"time"
	"xconn/common"
	"xconn/tools"
)

/**
 * @brief: 连接
 */
type UdpConn struct {
	common.BaseConn
	UdpAddr      *net.UDPAddr         // udp地址
	Conn         *net.UDPConn         // 连接
}



func newUdpConn(conn *net.UDPConn, addr *net.UDPAddr, config *common.Config)*UdpConn {
	localAddr := conn.LocalAddr().String()
	addrstr := addr.String()
	glog.Infoln(addrstr, localAddr)
	ci := &UdpConn{
		Conn:    conn,
		UdpAddr: addr,
	}
	ci.Id = uuid.New().String()
	ci.RemoteAddress = addr.String()
	ci.LocalAddr = conn.LocalAddr().String()
	ci.Sender = tools.NewDataTransport(1, config.SendChanSize)
	ci.Done = make(chan bool, 1)
	ci.TimeoutCheck = tools.NewTimeoutCheck(config.Interval, config.Timeout)
	if config.BufSize <= 0{
		config.BufSize = 1024
	}
	ci.RecvBufSize = config.BufSize
	ci.ConnCallback = config.ConnCallback
	ci.Label = config.Label
	ci.DataHandler = config.DataHandler
	ci.IConn = ci

	return ci
}

func (cl *UdpConn)Start(){
	go func() {
		defer func() {
			// 直接关闭
			cl.Close()
		}()

		cl.startSendProcess()

		if cl.ConnCallback != nil {
			// 新连接回调
			cl.ConnCallback.OnConnected(cl)
		}

		<-cl.Done

		if cl.ConnCallback != nil {
			// 关闭回调
			cl.ConnCallback.OnDisconnected(cl)
		}
	}()
}

/**
 * @brief: 发送处理流程
 */
func (cl *UdpConn)startSendProcess() {
	cl.Sender.Consume(func(data interface{}) bool {
		if data != nil{
			if bytess, ok := data.([]byte); ok{
				cl.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				if _, err := cl.Conn.WriteToUDP(bytess, cl.UdpAddr); err != nil {
					glog.Errorln("conn.Write", err.Error())
					cl.Done <- true
					return false
				}
			}
		}
		return true
	})
}

func (cl *UdpConn)recv(data []byte){
	cl.TimeoutCheck.Tick()

	if cl.DataHandler != nil{
		// udp
		cl.DataHandler.Handle(data, cl)
	}else{
		glog.Errorln("udp conn data handler is nil")
	}
}
