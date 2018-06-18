package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(minerAddress, nodeID string) {
	fmt.Printf("开启一个节点 %s\n",nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("正在挖矿,地址:", minerAddress)
		} else {
			log.Panic("错误的挖矿地址")
		}
	}
	StartServer(nodeID, minerAddress)
}