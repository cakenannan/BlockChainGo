package main

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
)

//定义区块
type Block struct {
	Timestamp int64		//时间线
	Transactions []*Transaction	//交易的集合
	PrevHash []byte		//上一个区块hash
	Hash []byte			//当前区块hash
	Nonce int			//随机数
}

//创建区块
func NewBlock(transactions []*Transaction, prevHash []byte) *Block {
	//block是一个指针,取得一个对象初始化后的地址
	block := &Block{time.Now().Unix(), transactions,prevHash,[]byte{}, 0}
	pow := NewProofOfWork(block)	//创建一个工作量证明的挖矿对象
	nonce,hash := pow.Run()		//开始挖矿
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//创建创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//对交易集合hash计算
func (block *Block) HashTransactions() []byte {
	var transactions [][]byte
	for _, tx := range block.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	return mTree.RootNode.data
}

//对象转为二进制字节集,写入文件
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//读取文件,二进制字节集转为对象
func DeserializeBlock(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}