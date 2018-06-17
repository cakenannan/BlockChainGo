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
	bc := NewBlockChain()
	utxoset := UTXOSet{bc}
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, &utxoset)	//转账交易
	cbTx := NewCoinBaseTX(from, "")					//奖励
	txs := []*Transaction{cbTx, tx}
	newblock := bc.MineBlock(txs)							//挖矿确认交易,记账成功
	utxoset.Update(newblock)
	fmt.Println("交易成功")
}
