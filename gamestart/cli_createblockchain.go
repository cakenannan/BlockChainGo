package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockChain(address string) {
	if !ValidateAddress(address) {
		log.Panic("地址不正确")
	}
	bc := CreateBlockChain(address) //创建区块链
	bc.db.Close()
	fmt.Println("创建成功", address)
}