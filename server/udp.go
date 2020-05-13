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
	baseConn
	udpAddr      *net.UDPAddr         // udp地址
	conn         *net.UDPConn         // 连接
}



func newUdpConn(conn *net.UDPConn, addr *net.UDPAddr, config *common.Config)*UdpConn {
	localAddr := conn.LocalAddr().String()
	addrstr := addr.String()
	glog.Infoln(addrstr, localAddr)
	ci := &UdpConn{
		conn:    conn,
		udpAddr: addr,
	}
	ci.id = uuid.New().String()
	ci.remoteAddress = addr.String()
	ci.localAddr = conn.LocalAddr().String()
	ci.sender = tools.NewDataTransport(1, config.SendChanSize)
	ci.receiver = tools.NewDataTransport(1, config.RecvChanSize)
	ci.done = make(chan bool, 1)
	ci.timeoutCheck = tools.NewTimeoutCheck(config.Interval, config.Timeout)
	if config.BufSize <= 0{
		config.BufSize = 1024
	}
	ci.recvBufSize = config.BufSize
	ci.connCallback = config.ConnCallback
	ci.label = config.Label
	ci.dataSplitter = config.DataSplitter
	ci.packetHandler = config.PacketHandler
	ci.iconn = ci

	return ci
}

func (cl *UdpConn)Start(){
	go func() {
		defer func() {
			// 直接关闭
			cl.Close()
		}()

		cl.startDataProcess()
		cl.startSendProcess()

		if cl.connCallback != nil {
			// 新连接回调
			cl.connCallback.OnConnected(cl)
		}

		<-cl.done

		if cl.connCallback != nil {
			// 关闭回调
			cl.connCallback.OnDisconnected(cl)
		}
	}()
}

/**
 * @brief: 发送处理流程
 */
func (cl *UdpConn)startSendProcess() {
	cl.sender.Consume(func(data interface{}) bool {
		if data != nil{
			if bytess, ok := data.([]byte); ok{
				cl.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				if _, err := cl.conn.WriteToUDP(bytess, cl.udpAddr); err != nil {
					glog.Errorln("conn.Write", err.Error())
					cl.done <- true
					return false
				}
			}
		}
		return true
	})
}
