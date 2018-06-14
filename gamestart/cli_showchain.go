package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) ShowBlockChain() {
	bc := NewBlockChain("")
	defer bc.db.Close()
	bci := bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("上一区块hash%x\n", block.PrevHash)
		fmt.Printf("当前区块hash:%x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s \n",strconv.FormatBool(pow.Validate()))
		for tx := range block.Transactions {
			fmt.Println("交易", tx)
		}
		fmt.Println()
		if len(block.PrevHash) == 0 {	//循环到创世区块,终止
			break
		}
	}
}
