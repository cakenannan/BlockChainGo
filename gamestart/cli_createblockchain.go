package main

import (
	"log"
	"fmt"
)

func (cli *CLI) createBlockChain(address string) {
	if !ValidateAddress(address) {
		log.Panic("地址不正确")
	}
	bc := CreateBlockChain(address) //创建区块链
	defer bc.db.Close()
	utxoset := UTXOSet{bc}
	utxoset.Reindex()				//在区块链创建之后,重建索引
	fmt.Println("创建成功", address)
}