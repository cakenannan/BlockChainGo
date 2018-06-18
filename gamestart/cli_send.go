package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("from地址不正确")
	}
	if !ValidateAddress(to) {
		log.Panic("to地址不正确")
	}
	bc := NewBlockChain(nodeID)
	utxoset := UTXOSet{bc}
	defer bc.db.Close()

	wallets,err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	tx := NewUTXOTransaction(&wallet, to, amount, &utxoset)	//转账交易
	if mineNow {	//立马挖矿
		cbTx := NewCoinBaseTX(from, "")					//奖励
		txs := []*Transaction{cbTx, tx}
		newblock := bc.MineBlock(txs)							//挖矿确认交易,记账成功
		utxoset.Update(newblock)
	} else {
		sendTx(knowNodes[0],tx)			//发送交易等待确认
	}

	fmt.Println("交易成功")
}
