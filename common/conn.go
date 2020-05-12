package common

import (
	"github.com/gin-gonic/gin"
	"time"
)

/**
 * 数据包获取接口，跟业务无关
 */
type Getter interface {
	/**
	 * @brief: 包解析器接口
	 * @param1 bytess: 当前接收到的数据
	 * @param2 conn: 当前conn
	 * @return1: 获取出来包内容，多个
	 * @return2: 剩余数据
	 * @return3: 错误信息
	 */
	Get([]byte, IConn)([][]byte, []byte, error)
}

/**
 * 包解析器
 */
type Parser interface {
	/**
	 * @brief: 包解析器
	 * @param1 bytess: 包数据
	 * @param2 tcpConn: 当前连接
	 */
	Parse([]byte, IConn)
}

/**
 * 连接回调接口
 */
type ConnCallback interface {
	/**
	 * @brief: 新连接回调
	 * @param1 conn: 连接
	 */
	OnConnected(conn IConn)               //
	/**
	 * @brief: 连接断开回调
	 * @param1 conn: 连接
	 * @param2 err: 错误信息，正常关闭的为nil
	 */
	OnDisconnected(conn IConn) //
	/**
	 * @brief: 错误回调
	 * @param1 conn: 连接
	 * @param2 err: 错误信息
	 */
	OnError(conn IConn, err error)
}

/**
 * tcp 配置信息
 */
type Config struct {
	Ip           string            // ip
	Port         int               // 端口
	Network      string            // 默认tcp, 另外可以有tcp4, tcp6, "unix" or "unixpacket", udp, ws(websocket）
	Interval     time.Duration     // 心跳间隔
	Timeout      time.Duration     // 超时时间
	BufSize      int               // 接收缓冲区大小
	SendChanSize int               // 发送通道大小
	RecvChanSize int               // 接收通道大小
	Getter       Getter            // tcp 连接处理接口
	Parser       Parser            // 包解析器
	ConnCallback ConnCallback      // 连接回调接口
	Label        string            // 标签
	WsUrls       map[string]string // key: path, value: text(或者binary), 当Type为ws有效
	WsGin        *gin.Engine       // websocket 对应的gin engine对象，当Type为ws有效
}

/**
 * @brief: 连接接口
 */
type IConn interface {
	Start()
	Close()
	Send([]byte)
	GetId()string
	GetTag(string)interface{}
	SetTag(string, interface{})
	GetLabel()string
	SetLabel(string)
	GetRemoteAddr()string
	GetLocalAddr()string
}
