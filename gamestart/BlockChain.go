package main

type BlockChain struct {
	blocks []*Block		//区块数组,元素是指针,即区块地址
}

//增加一个区块
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks) - 1]	//取出最后一个区块
	newBlock := NewBlock(data, prevBlock.Hash)		//创建区块
	chain.blocks = append(chain.blocks, newBlock)	//区块链插入新区块
}

//创建区块链
func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{NewGenesisBlock()}}	//包含创世区块
}
