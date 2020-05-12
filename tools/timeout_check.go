package tools

import (
	"context"
	"github.com/golang/glog"
	"time"
)

type TimeoutCheck struct {
	timer        *time.Ticker  // 定时器
	interval     time.Duration // 间隔
	timeout      time.Duration // 超时时间长度
	lastTickTime time.Time     // 最后tick 时间
	ctx       context.Context	// 上下文
	cancel    context.CancelFunc // cancel 函数
}

/**
 * @brief: 创建timeoutcheck
 * @param1 interval: 时间间隔
 * @param2 timeout: 超时时间
 */
func NewTimeoutCheck(interval, timeout time.Duration)*TimeoutCheck {
	tc := &TimeoutCheck{}
	tc.interval = interval
	tc.timeout = timeout

	tc.ctx, tc.cancel = context.WithCancel(context.Background())

	return tc
}

/**
 * @brief: 取消
 */
func (tc *TimeoutCheck)Cancel(){
	if tc.cancel != nil{
		tc.cancel()
	}
}

/**
 * @brief: 设置最后tick时间
 */
func (tc *TimeoutCheck)Tick(){
	tc.lastTickTime = time.Now()
}

/**
 * @brief: 检测是否过期
 * @param1 cb: 回调函数
 */
func (tc *TimeoutCheck)Check(cb func(bool)){
	if cb == nil{
		glog.Errorln("TimeoutCheck.Check参数为nil")
		return
	}

	go func(){
		if int(tc.interval) <= 0{
			glog.Infoln("时间间隔小于等于0，不启动超时检测定时器")
			return
		}

		tc.timer = time.NewTicker(tc.interval)
		defer func(){
			tc.timer.Stop()
		}()

		for{
			select {
			case <-tc.ctx.Done():
				glog.Infoln( "Check ctx.Done")
				return
			case <-tc.timer.C:
				now := time.Now()
				if now.Sub(tc.lastTickTime) > tc.timeout{
					glog.Errorln("heartbeat timeout exit", now.Format("2006-01-02 15:04:05"), tc.lastTickTime.Format("2006-01-02 15:04:05"))
					cb(true)
					return
				}
			}
		}
	}()
}
