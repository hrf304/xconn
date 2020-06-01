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
type TcpConn struct {
	common.BaseConn
	Conn         net.Conn             // 连接
}

func newTcpConn(conn net.Conn, config *common.Config)*TcpConn {
	ci := &TcpConn{
		Conn:  conn,
	}
	ci.Id = uuid.New().String()
	ci.RemoteAddress = conn.RemoteAddr().String()
	ci.LocalAddr = conn.LocalAddr().String()
	ci.Sender = tools.NewDataTransport(1, config.SendChanSize)
	ci.Done = make(chan bool, 1)
	ci.TimeoutCheck = tools.NewTimeoutCheck(config.Interval, config.Timeout)
	if config.BufSize <= 0{
		config.BufSize = 1024
	}
	ci.RecvBufSize = config.BufSize
	ci.ConnCallback = config.ConnCallback
	ci.DataHandler = config.DataHandler
	ci.Label = config.Label
	ci.IConn = ci

	return ci
}

func (cl *TcpConn)Start(){
	go func() {
		defer func() {
			// 直接关闭
			cl.Close()
		}()

		cl.startSendProcess()
		cl.startRecvProcess()

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

func (cl *TcpConn)Close(){
	cl.Close()

	if cl.Conn != nil{
		cl.Conn.Close()
	}
}

/**
 * @brief: 发送处理流程
 */
func (cl *TcpConn)startSendProcess() {
	cl.Sender.Consume(func(data interface{}) bool {
		if data != nil{
			if bytess, ok := data.([]byte); ok{
				cl.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				if _, err := cl.Conn.Write(bytess); err != nil {
					glog.Errorln("conn.Write", err.Error())
					cl.Done <- true
					return false
				}
			}
		}
		return true
	})
}


/**
 * @brief: 接收处理流程
 */
func (cl *TcpConn)startRecvProcess(){
	go func() {
		defer func() {
			cl.Done <- true
		}()

		recvBuffer := make([]byte, cl.RecvBufSize)
		ringBuf := tools.NewRingBuffer(65535)
		for {
			i, err := cl.Conn.Read(recvBuffer) // 读取数据
			if err != nil {
				glog.Errorln(cl.Label, "读取客户端数据错误:", err.Error())
				if cl.ConnCallback != nil{
					// 新连接回调
					cl.ConnCallback.OnError(cl, err)
				}
				break
			}
			i, err = ringBuf.Write(recvBuffer[0:i])
			if err != nil {
				glog.Errorln(cl.Label, "写入数据到本地换成buf错误:", err.Error())
				break
			}

			// 处理数据
			cl.TimeoutCheck.Tick()

			// handle data
			if cl.DataHandler != nil {
				left, err := cl.DataHandler.Handle(ringBuf.Bytes(), cl)
				if err != nil {
					glog.Errorln("getter get err", err.Error())
					ringBuf.Reset()
					continue
				}

				// 未处理完的重新写入
				ringBuf.Reset()
				if left != nil {
					ringBuf.Write(left)
				}
			}else{
				glog.Errorln("data handler is nil")
			}
		}
	}()
}
