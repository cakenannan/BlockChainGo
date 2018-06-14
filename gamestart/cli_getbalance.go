package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("地址不正确")
	}
	bc := NewBlockChain(address)
	defer bc.db.Close()
	balance := 0
	pubkeyhash := Base58Decode([]byte(address))
	pubkeyhash = pubkeyhash[1:len(pubkeyhash)-4]
	UTXOs := bc.FindUTXO(pubkeyhash)	//查找交易金额
	for _, out := range UTXOs {
		balance += out.Value	//取出金额
	}
	fmt.Printf("查询的地址为:%s,金额为:%d \n", address, balance)
}
