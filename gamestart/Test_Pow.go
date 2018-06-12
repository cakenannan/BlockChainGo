package main

import (
	"crypto/sha256"
	"strconv"
	"fmt"
	"time"
)

func mainx() {
	start := time.Now()
	for i:=0; i<100000000;i++{
		data := sha256.Sum256([]byte(strconv.Itoa(i)))
		fmt.Printf("%10d,%x\n", i, data)
		//fmt.Printf("%s\n", string(data[len(data)-1:]))
		if string(data[len(data)-2:]) == "00" {
			usedTime := time.Since(start)
			fmt.Printf("挖矿成功i=%d,time=%dMs", i, usedTime)
			break
		}
	}
}
