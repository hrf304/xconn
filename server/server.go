package server

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
	"xconn/common"
)

/**
 * @brief: 服务端对象
 */
type Server struct {
	config       *common.Config      // 配置
	connMap      sync.Map     // 连接列表,ip:port为key, conn为value
	connCallback common.ConnCallback // 回调函数

	// websocket相关
	upgrader websocket.Upgrader
}

func NewServer(config *common.Config)*Server {
	if config == nil {
		glog.Error("参数config为nil")
		return nil
	}

	s := &Server{
		config:       config,
		connCallback: config.ConnCallback,
	}
	config.ConnCallback = s

	if config.Network == "ws" {
		// websocket额外设置
		s.upgrader = websocket.Upgrader{
			ReadBufferSize:   4096,
			WriteBufferSize:  4096,
			HandshakeTimeout: 5 * time.Second,
			// 取消ws跨域校验
			CheckOrigin: func(r *http.Request) bool {
				return true
			}}
	}
	return s
}

func (ts *Server)Start() {
	if ts.config.Network == "tcp" || ts.config.Network == "tcp4" || ts.config.Network == "tcp6" || ts.config.Network == "unix" || ts.config.Network == "unixpacket" {
		ts.startTcpServer()
	} else if ts.config.Network == "udp" {
		ts.startUdpServer()
	} else if ts.config.Network == "ws" {
		ts.startWsServer()
	}
}

/**
 * @brief: 启动TCP服务端
 */
func (ts *Server)startTcpServer(){
	network := ts.config.Network
	if network == ""{
		network = "tcp"
	}

	listen, err := net.Listen(network, ts.config.Ip + ":" + strconv.Itoa(ts.config.Port))
	if err != nil {
		glog.Errorln("监听端口失败：", err.Error())
		return
	}

	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				glog.Errorln("接受TCP连接异常:", err.Error())
				continue
			}
			glog.Infoln("TCP连接来自:", conn.RemoteAddr().String())

			go func(conn net.Conn, config *common.Config){
				iconn := newTcpConn(conn, config)
				iconn.Start()
			}(conn, ts.config)
		}
	}()

	return
}

/**
 * @brief: 启动UDP服务端
 */
func (ts *Server)startUdpServer(){
	addr, err := net.ResolveUDPAddr("udp", ts.config.Ip + ":" + strconv.Itoa(ts.config.Port))
	if err != nil{
		glog.Errorln(err.Error())
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil{
		glog.Errorln(err.Error())
		return
	}


	go func() {
		buf := make([]byte, 65535)
		for {
			n, radd, err := conn.ReadFromUDP(buf)
			if err != nil {
				glog.Errorln(err.Error())
				continue
			}
			if n <= 0 {
				continue
			}

			copyBuf := copyBytes(buf[:n])
			if v, ok := ts.connMap.Load(radd.String()); ok {
				if ccon, ok1 := v.(*UdpConn); ok1 {
					ccon.recv(copyBuf)
				}
			} else {
				ccon := newUdpConn(conn, radd, ts.config)
				ccon.Start()
				ccon.recv(copyBuf)
			}
		}
	}()
}

/**
 * @brief: 启动ws服务端
 */
func (ts *Server)startWsServer(){
	for path, wsMsgType := range ts.config.WsUrls{
		go func(p, mt string){
			ts.config.WsGin.GET(p, func(ctx *gin.Context) {
				conn, err := ts.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
				if err != nil {
					glog.Error(err)
					ctx.JSON(500, gin.H{"code": 500, "msg": err.Error(), "data": nil})
					return
				}

				wsconn := newWsConn(conn, ctx, ts.config, p, mt)
				wsconn.Start()
			})
		}(path, wsMsgType)
	}
}

/**
 * @brief: 广播
 */
func (ts *Server)GetAllConn()[]common.IConn{
	cons := []common.IConn{}
	ts.connMap.Range(func(key, value interface{}) bool {
		if con, ok := value.(common.IConn); ok{
			cons = append(cons, con)
		}
		return true
	})

	return cons
}

/**
 * @brief: 新连接回调
 * @param1 conn: 连接
 */
func (ts *Server)OnConnected(conn common.IConn){
	if conn == nil{
		return
	}

	ts.connMap.Store(conn.GetRemoteAddr(), conn)

	if ts.connCallback != nil{
		ts.connCallback.OnConnected(conn)
	}
}

/**
 * @brief: 连接断开回调
 * @param1 conn: 连接
 * @param2 err: 错误信息，正常关闭的为nil
 */
func (ts *Server)OnDisconnected(conn common.IConn){
	if conn == nil{
		return
	}

	ts.connMap.Delete(conn.GetRemoteAddr())

	if ts.connCallback != nil{
		ts.connCallback.OnDisconnected(conn)
	}
}

/**
 * @brief: 错误回调
 * @param1 conn: 连接
 * @param2 err: 错误信息
 */
func (ts *Server)OnError(conn common.IConn, err error){
	if ts.connCallback != nil{
		ts.connCallback.OnError(conn, err)
	}
}

func copyBytes(buf []byte)[]byte{
	copyBuf := make([]byte, len(buf))
	copy(copyBuf, buf)

	return copyBuf
}
