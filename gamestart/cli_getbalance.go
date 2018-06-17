package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("地址不正确")
	}
	bc := NewBlockChain()
	utxoset := UTXOSet{bc}
	defer bc.db.Close()
	balance := 0
	pubkeyhash := Base58Decode([]byte(address))
	pubkeyhash = pubkeyhash[1:len(pubkeyhash)-4]
	UTXOs := utxoset.FindUTXO(pubkeyhash)	//查找所有未花费输出
	for _, out := range UTXOs {
		balance += out.Value	//累加金额
	}
	fmt.Printf("查询的地址为:%s,金额为:%d \n", address, balance)
}
