package main

import (
	"bytes"
	"encoding/binary"
	"github.com/labstack/gommon/log"
)

//整数转化为十六进制
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)	//开辟内存,存储字节集
	err := binary.Write(buff, binary.BigEndian, num)	//num转化字节集写入
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()		//返回字节集合
}

//将字节集合反转
func ReverseBytes(data []byte) {
	for i,j := 0, len(data)-1; i < j; i,j = i+1,j-1 {
		data[i],data[j] = data[j],data[i]
	}
}
