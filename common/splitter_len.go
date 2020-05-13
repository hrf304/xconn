package common

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"strconv"
)

var _lenCalculator map[string]func([]byte)int64

func init(){
	_lenCalculator = make(map[string]func([]byte)int64)
	_lenCalculator["1_1"] = func(bytes []byte) int64 {
		// 大端1位
		return int64(bytes[0])
	}
	_lenCalculator["1_2"] = func(bytes []byte) int64 {
		// 大端2位
		s := binary.BigEndian.Uint16(bytes)
		return int64(s)
	}
	_lenCalculator["1_4"] = func(bytes []byte) int64 {
		// 大端4位
		s := binary.BigEndian.Uint32(bytes)
		return int64(s)
	}
	_lenCalculator["1_8"] = func(bytes []byte) int64 {
		// 大端8位
		s := binary.BigEndian.Uint64(bytes)
		return int64(s)
	}
	_lenCalculator["0_1"] = func(bytes []byte) int64 {
		// 小端1位
		return int64(bytes[0])
	}
	_lenCalculator["0_2"] = func(bytes []byte) int64 {
		// 小端2位
		s := binary.LittleEndian.Uint16(bytes)
		return int64(s)
	}
	_lenCalculator["0_4"] = func(bytes []byte) int64 {
		// 小端4位
		s := binary.LittleEndian.Uint32(bytes)
		return int64(s)
	}
	_lenCalculator["0_8"] = func(bytes []byte) int64 {
		// 小端8位
		s := binary.LittleEndian.Uint64(bytes)
		return int64(s)
	}
}

/**
 * @brief: 以数据包长度开始数据拆分
 */
type LenSplitter struct {
	lenByteCount   int64  // 数据包长度字节数，有1，2，4，8
	isBigEndian    bool // 是否大端
	lenFuncKey string 	// 数据包长度计算接口key  例如：1_1
	containLenByte bool // 数据包长度是否包含数据包长度字节数
}

/**
 * 拆分数据包
 */
func (ls *LenSplitter)Split(data []byte, con IConn)([][]byte, []byte, error){
	if len(data) < int(ls.lenByteCount){
		// 不完整数据包，直接返回
		return nil, data, nil
	}

	ps := make([][]byte, 0)
	index := 0

	for {
		temp := data[index:]
		var dlen int64 = 0
		if f, ok := _lenCalculator[ls.lenFuncKey]; ok {
			dlen = f(temp[:ls.lenByteCount])
			if ls.containLenByte {
				// 如果包含长度几个字节，则减去
				dlen = dlen - ls.lenByteCount
			}
		} else {
			// key不存在，直接报错
			return nil, nil, errors.New("unknown len calculator " + ls.lenFuncKey)
		}
		if int64(len(temp))-ls.lenByteCount < dlen {
			// 不完整数据包，直接返回
			return ps, temp, nil
		}
		ps = append(ps, temp[ls.lenByteCount:ls.lenByteCount+dlen])
		index += int(ls.lenByteCount + dlen)
	}

	return ps, data[index:], nil
}

/**
 * @brief: 创建以数据包长度开始数据拆分
 */
func NewLenSplitter(lenByteCount int, isBigEndian, containLenByte bool)*LenSplitter{
	if lenByteCount != 1 && lenByteCount != 2 && lenByteCount != 4 && lenByteCount != 8{
		return nil
	}
	sp := &LenSplitter{}
	sp.lenByteCount = int64(lenByteCount)
	sp.isBigEndian = isBigEndian
	if isBigEndian{
		sp.lenFuncKey = "1_" + strconv.Itoa(lenByteCount)
	}else{
		sp.lenFuncKey = "0_" + strconv.Itoa(lenByteCount)
	}
	sp.containLenByte = containLenByte

	return sp
}
