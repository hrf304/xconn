package client

import (
	"fmt"
	"github.com/golang/glog"
	"net"
	"sync"
	"time"
	"xconn/common"
)

type Client struct{
	config       *common.Config      // 配置
	connMap      sync.Map     // 连接列表,ip:port为key, conn为value
	connCallback common.ConnCallback // 回调函数
}

func NewClient(config *common.Config)*Client{
	if config == nil {
		glog.Error("参数config为nil")
		return nil
	}

	c := &Client{
		config:       config,
		connCallback: config.ConnCallback,
	}

	return c
}

func (c *Client)Dial(){
	if c.config.Network == "tcp" || c.config.Network == "tcp4" || c.config.Network == "tcp6" || c.config.Network == "unix" || c.config.Network == "unixpacket" {
		c.dialTcpServer()
	} else if c.config.Network == "udp" {
		c.dialUdpServer()
	} else if c.config.Network == "ws" {
		glog.Error("can not dial websocket")
	}
}

func (c *Client)dialTcpServer(){
	address := fmt.Sprintf("%s:%d", c.config.Ip, c.config.Port)
	addr, err := net.ResolveTCPAddr(c.config.Network, address)
	if err != nil{
		glog.Errorln("resolve tcp addr error", err.Error())
		return
	}

	con, err := net.DialTCP(c.config.Network, nil, addr)
	if err != nil{
		glog.Errorln("dial tcp error", err.Error())
		return
	}

	c.config.ConnCallback = c
	iconn := newTcpConn(con, c.config)
	iconn.Start()
}

func (c *Client)dialUdpServer(){
	address := fmt.Sprintf("%s:%d", c.config.Ip, c.config.Port)
	addr, err := net.ResolveUDPAddr(c.config.Network, address)
	if err != nil{
		glog.Errorln("resolve udp addr error", err.Error())
		return
	}

	con, err := net.DialUDP(c.config.Network, nil, addr)
	if err != nil{
		glog.Errorln("dial tcp error", err.Error())
		return
	}

	c.config.ConnCallback = c
	iconn := newUdpConn(con, addr, c.config)
	iconn.Start()
}


/**
 * @brief: 新连接回调
 * @param1 conn: 连接
 */
func (c *Client)OnConnected(conn common.IConn){
	if conn == nil{
		return
	}

	c.connMap.Store(conn.GetRemoteAddr(), conn)

	if c.connCallback != nil{
		c.connCallback.OnConnected(conn)
	}
}

/**
 * @brief: 连接断开回调
 * @param1 conn: 连接
 * @param2 err: 错误信息，正常关闭的为nil
 */
func (c *Client)OnDisconnected(conn common.IConn){
	if conn == nil{
		return
	}

	c.connMap.Delete(conn.GetRemoteAddr())

	if c.connCallback != nil{
		c.connCallback.OnDisconnected(conn)
	}
}

/**
 * @brief: 错误回调
 * @param1 conn: 连接
 * @param2 err: 错误信息
 */
func (c *Client)OnError(conn common.IConn, err error){
	if c.connCallback != nil{
		c.connCallback.OnError(conn, err)
	}
}
