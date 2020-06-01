package server

import (
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"time"
	"xconn/common"
	"xconn/tools"
)

/**
 * @brief: 连接
 */
type WsConn struct {
	common.BaseConn
	conn         *websocket.Conn      // 连接
	ginCtx       *gin.Context         // 当前连接上下文
	path         string               // 连接对应路径
	msgType      int                  // 消息类型
}


func newWsConn(conn *websocket.Conn, ctx *gin.Context, config *common.Config, path, wsMsgType string)*WsConn {
	msgType := 0
	if wsMsgType == "text" {
		msgType = websocket.TextMessage
	} else {
		msgType = websocket.BinaryMessage
	}
	ci := &WsConn{
		conn:    conn,
		ginCtx:  ctx,
		path:    path,
		msgType: msgType,
	}
	ci.Id = uuid.New().String()
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
	ci.RemoteAddress = conn.RemoteAddr().String()
	ci.LocalAddr = conn.LocalAddr().String()
	ci.IConn = ci

	return ci
}

/**
 * 启动
 */
func (cl *WsConn)Start() {
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

func (cl *WsConn)Close(){
	cl.Close()

	if cl.conn != nil{
		cl.conn.Close()
	}
}

/**
 * @brief: 获取查询参数值
 * @param1 key: 参数名称
 * @return1: 返回值
 */
func (cl *WsConn)GetQuery(key string)string{
	if key == "" || cl.ginCtx == nil{
		return ""
	}
	return cl.ginCtx.DefaultQuery(key, "")
}

/**
 * 获取连接对应的path
 */
func (cl *WsConn)GetPath()string{
	return cl.path
}

/**
 * @brief: 发送处理流程
 */
func (cl *WsConn)startSendProcess() {
	cl.Sender.Consume(func(data interface{}) bool {
		if data != nil{
			if bytess, ok := data.([]byte); ok{
				cl.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				if err := cl.conn.WriteMessage(cl.msgType, bytess); err != nil {
					glog.Errorln("conn.Write", err.Error())
					cl.Done <- true
					return false
				}else{
					//glog.Infoln("-------------->发送成功")
				}
			}
		}
		return true
	})
}


/**
 * @brief: 接收处理流程
 */
func (cl *WsConn)startRecvProcess() {
	go func() {
		defer func() {
			cl.Done <- true
		}()

		for {
			_, data, err := cl.conn.ReadMessage() // 读取数据
			if err != nil {
				glog.Errorln(cl.Label, "读取客户端数据错误:", err.Error())
				if cl.ConnCallback != nil {
					// 新连接回调
					cl.ConnCallback.OnError(cl, err)
				}
				break
			}

			// 处理数据
			cl.TimeoutCheck.Tick()
			// websocket 不需要处理粘包问题
			if cl.DataHandler != nil{
				cl.DataHandler.Handle(data, cl)
			}else{
				glog.Errorln("udp conn data handler is nil")
			}
		}
	}()
}

