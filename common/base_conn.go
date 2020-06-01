package common

import (
	"github.com/golang/glog"
	"sync"
	"xconn/tools"
)

/**
 * @brief: 连接基础
 */
type BaseConn struct {
	Id            string               // id
	RemoteAddress string               // 地址
	LocalAddr     string               // 本地地址
	Sender        *tools.DataTransport // 发送队列
	TimeoutCheck  *tools.TimeoutCheck  // 超时检测
	Done          chan bool            // 标识是否完成
	RecvBufSize   int                  // 接收缓冲区大小
	ConnCallback  ConnCallback         // 服务端
	DataHandler DataHandler        // 包解析器
	Label         string               // 标签
	Tag           sync.Map             // 自定义数据
	IConn         IConn
}

/**
 * 获取地址
 */
func (cl *BaseConn)GetId()string{
	return cl.Id
}

func (cl *BaseConn)Send(data []byte){
	if data == nil{
		glog.Errorln("发送数据位nil")
		return
	}

	cl.Sender.Produce(data)
}

func (cl *BaseConn)Close(){
	cl.TimeoutCheck.Cancel()
	cl.Sender.Cancel()
}

func (cl *BaseConn)GetTag(key string)interface{}{
	if key == "" {
		return nil
	}
	if v, ok := cl.Tag.Load(key); !ok{
		return nil
	}else{
		return v
	}
}

func (cl *BaseConn)SetTag(key string, tag interface{}){
	cl.Tag.Store(key, tag)
}

func (cl *BaseConn)GetLabel()string{
	return cl.Label
}

func (cl *BaseConn)SetLabel(label string) {
	cl.Label = label
}

/**
 * 获取地址
 */
func (cl *BaseConn)GetRemoteAddr()string{
	return cl.RemoteAddress
}

/**
 * 获取本地地址
 */
func (cl *BaseConn)GetLocalAddr()string{
	return cl.LocalAddr
}

/**
 * @brief: 超时检测进程
 */
func (cl *BaseConn)StartTimeoutCheckProcess() {
	cl.TimeoutCheck.Check(func(b bool) {
		if b {
			cl.Done <- true
		}
	})
}
