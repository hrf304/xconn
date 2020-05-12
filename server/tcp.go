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
	baseConn
	conn         net.Conn             // 连接
}

func newTcpConn(conn net.Conn, config *common.Config)*TcpConn {
	ci := &TcpConn{
		conn:  conn,
	}
	ci.id = uuid.New().String()
	ci.remoteAddress = conn.RemoteAddr().String()
	ci.localAddr = conn.LocalAddr().String()
	ci.sender = tools.NewDataTransport(1, config.SendChanSize)
	ci.receiver = tools.NewDataTransport(1, config.RecvChanSize)
	ci.done = make(chan bool, 1)
	ci.timeoutCheck = tools.NewTimeoutCheck(config.Interval, config.Timeout)
	ci.recvBufSize = config.BufSize
	ci.connCallback = config.ConnCallback
	ci.getter = config.Getter
	ci.parser = config.Parser
	ci.label = config.Label
	ci.iconn = ci

	return ci
}

func (cl *TcpConn)Start(){
	go func() {
		defer func() {
			// 直接关闭
			cl.Close()
		}()

		cl.startDataProcess()
		cl.startSendProcess()
		cl.startRecvProcess()

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

func (cl *TcpConn)Close(){
	cl.baseConn.Close()

	if cl.conn != nil{
		cl.conn.Close()
	}
}

/**
 * @brief: 发送处理流程
 */
func (cl *TcpConn)startSendProcess() {
	cl.sender.Consume(func(data interface{}) bool {
		if data != nil{
			if bytess, ok := data.([]byte); ok{
				cl.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				if _, err := cl.conn.Write(bytess); err != nil {
					glog.Errorln("conn.Write", err.Error())
					cl.done <- true
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
			cl.done <- true
		}()

		recvBuffer := make([]byte, cl.recvBufSize)
		ringBuf := tools.NewRingBuffer(65535)
		for {
			i, err := cl.conn.Read(recvBuffer) // 读取数据
			if err != nil {
				glog.Errorln(cl.label, "读取客户端数据错误:", err.Error())
				if cl.connCallback != nil{
					// 新连接回调
					cl.connCallback.OnError(cl, err)
				}
				break
			}
			i, err = ringBuf.Write(recvBuffer[0:i])
			if err != nil {
				glog.Errorln(cl.label, "写入数据到本地换成buf错误:", err.Error())
				break
			}

			// 处理数据
			cl.timeoutCheck.Tick()

			if cl.getter != nil {
				ps, left, err := cl.getter.Get(ringBuf.Bytes(), cl)
				if err != nil {
					glog.Errorln("getter get err", err.Error())
					ringBuf.Reset()
					continue
				}
				if ps != nil {
					for i := range ps {
						cl.receiver.Produce(ps[i])
					}
				}

				// 未处理完的重新写入
				ringBuf.Reset()
				if left != nil {
					ringBuf.Write(left)
				}
			}else{
				cl.receiver.Produce(ringBuf.Bytes())
				ringBuf.Reset()
			}
		}
	}()
}
