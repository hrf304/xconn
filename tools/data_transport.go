package tools

import (
	"context"
	"github.com/golang/glog"
)

/**
 * @brief: 数据传输
 */
type DataTransport struct {
	dataChans []chan interface{}
	ctx       context.Context
	cancel    context.CancelFunc
	index     int
}

/**
 * @brief: 创建数据传输
 * @param1 cc(consume count): 消费队列数量
 * @param2 cap: 数据队列容量
 */
func NewDataTransport(cc, cap int)*DataTransport{
	if cc <= 0{
		cc = 1
	}
	if cap <= 0 {
		cap = 1000
	}
	dt := &DataTransport{}
	dt.dataChans = make([]chan interface{}, cc)
	for i := range dt.dataChans{
		dt.dataChans[i] = make(chan interface{}, cap)
	}

	dt.index = 0
	ctx, cancel := context.WithCancel(context.Background())
	dt.ctx = ctx
	dt.cancel = cancel

	return dt
}

/**
 * @brief: 取消经常
 */
func (dt *DataTransport)Cancel(){
	if dt.cancel != nil{
		dt.cancel()
	}

	if dt.dataChans != nil{
		for i := range dt.dataChans{
			close(dt.dataChans[i])
		}
	}
}

/**
 * @brief: 数据生产
 * @param1 data: 数据，如果是指针类型，建议使用深拷贝模式创建新对象传入
 */
func (dt *DataTransport)Produce(data interface{}){
	if data == nil{
		return
	}

	defer func() {
		if x := recover(); x != nil {
			glog.Errorln("DataTransport.Produce recover: %v", x)
		}
	}()

	if len(dt.dataChans) == 1{
		dt.dataChans[0] <- data
	}else{
		// 按顺序分配给各个处理队列
		dt.dataChans[dt.index % len(dt.dataChans)] <- data
		if dt.index > 65535{
			dt.index = 0
		}else{
			dt.index++
		}
	}
}

/**
 * @brief: 数据消费
 * @param1 cb: 回调函数, 返回false可以终端整个消费流程
 */
func (dt *DataTransport)Consume(cb func(interface{})bool){
	if cb == nil{
		glog.Errorln("Consume参数为nil")
		return
	}

	for i := range dt.dataChans{
		go func(dc chan interface{}){
			defer func() {
				if x := recover(); x != nil {
					glog.Errorln("DataTransport.Consume recover: %v", x)
				}
			}()

			glog.Infoln("Consume 启动")
			for{
				select {
				case <-dt.ctx.Done():
					glog.Infoln("DataTransport.Consume ctx.Done")
					return
				case data := <-dc:
					if !cb(data){
						return
					}
				}
			}
		}(dt.dataChans[i])
	}
}
