package server

import (
	"github.com/golang/glog"
	"sync"
	"xconn/common"
	"xconn/tools"
)


type baseConn struct {
	id            string               // id
	remoteAddress string               // 地址
	localAddr     string               // 本地地址
	sender        *tools.DataTransport // 发送队列
	receiver      *tools.DataTransport // 接收队列
	timeoutCheck  *tools.TimeoutCheck  // 超时检测
	done          chan bool            // 标识是否完成
	recvBufSize   int                  // 接收缓冲区大小
	connCallback  common.ConnCallback  // 服务端
	getter        common.Getter        // 数据获取（分包器）
	parser        common.Parser        // 包解析器
	label         string               // 标签
	tag           sync.Map             // 自定义数据
	iconn         common.IConn
}

/**
 * 获取地址
 */
func (cl *baseConn)GetId()string{
	return cl.id
}

func (cl *baseConn)Send(data []byte){
	if data == nil{
		glog.Errorln("发送数据位nil")
		return
	}

	cl.sender.Produce(data)
}

/**
 * @brief:接收到数据
 */
func (cl *baseConn)Recv(data []byte) {
	// 处理数据
	cl.timeoutCheck.Tick()
	cl.receiver.Produce(data)
}

func (cl *baseConn)Close(){
	cl.timeoutCheck.Cancel()
	cl.sender.Cancel()
	cl.receiver.Cancel()
}

func (cl *baseConn)GetTag(key string)interface{}{
	if key == "" {
		return nil
	}
	if v, ok := cl.tag.Load(key); !ok{
		return nil
	}else{
		return v
	}
}

func (cl *baseConn)SetTag(key string, tag interface{}){
	cl.tag.Store(key, tag)
}

func (cl *baseConn)GetLabel()string{
	return cl.label
}

func (cl *baseConn)SetLabel(label string) {
	cl.label = label
}

/**
 * 获取地址
 */
func (cl *baseConn)GetRemoteAddr()string{
	return cl.remoteAddress
}

/**
 * 获取本地地址
 */
func (cl *baseConn)GetLocalAddr()string{
	return cl.localAddr
}

/**
 * @brief: 超时检测进程
 */
func (cl *baseConn)startTimeoutCheckProcess() {
	cl.timeoutCheck.Check(func(b bool) {
		if b {
			cl.done <- true
		}
	})
}
/**
 * @brief: 启动数据处理流程
 */
func (cl *baseConn)startDataProcess(){
	cl.receiver.Consume(func(data interface{}) bool {
		if data != nil{
			if bytess, ok := data.([]byte); ok{
				if cl.parser != nil{
					cl.parser.Parse(bytess, cl.iconn)
				}
			}
		}
		return true
	})
}
