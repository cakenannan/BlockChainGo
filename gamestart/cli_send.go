package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("from地址不正确")
	}
	if !ValidateAddress(to) {
		log.Panic("to地址不正确")
	}
	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)	//转账
	bc.MineBlock([]*Transaction{tx})	//挖矿确认交易,记账成功
	fmt.Println("交易成功")
}
