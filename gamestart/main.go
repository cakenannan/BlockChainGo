package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("game start")

	bc := NewBlockChain()	//创建区块链
	bc.AddBlock("第二个区块数据")
	bc.AddBlock("第三个区块数据")

	for _, block := range bc.blocks {
		fmt.Printf("上一区块hash%x\n", block.PrevHash)
		fmt.Printf("数据:%s\n", block.Data)
		fmt.Printf("当前区块hash:%x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s\n",strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
