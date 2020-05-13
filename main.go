package main

import (
	"github.com/golang/glog"
	"time"
	"xconn/common"
	"xconn/server"
)

func main() {
	config := &common.Config{}
	config.Network = "udp"
	config.RecvChanSize = 1000
	config.ConnCallback = &TestConnCallback{}
	config.WsUrls = nil
	config.SendChanSize = 1000
	config.Timeout = time.Second * 60
	config.Interval = time.Second * 5
	config.DataSplitter = nil		// udp 下getter不用
	config.Port = 5060
	config.Ip = ""
	config.Label = "gb28181"
	config.BufSize = 65535
	config.PacketHandler = &TestParser{}

	svr := server.NewServer(config)
	svr.Start()
}

type TestConnCallback struct{

}

/**
 * @brief: 新连接回调
 * @param1 conn: 连接
 */
func (cc *TestConnCallback)OnConnected(con common.IConn){
	glog.Infoln("OnConnected", con.GetLabel(), con.GetRemoteAddr() )
}

/**
 * @brief: 连接断开回调
 * @param1 conn: 连接
 * @param2 err: 错误信息，正常关闭的为nil
 */
func (cc *TestConnCallback)OnDisconnected(con common.IConn){
	glog.Infoln("OnDisconnected", con.GetLabel(), con.GetRemoteAddr() )
}

/**
 * @brief: 错误回调
 * @param1 conn: 连接
 * @param2 err: 错误信息
 */
func (cc *TestConnCallback)OnError(con common.IConn, err error){
	glog.Infoln("OnError", con.GetLabel(), con.GetRemoteAddr() )
}

type TestParser struct{

}

func (p *TestParser)Handle(data []byte, con common.IConn) {
}
