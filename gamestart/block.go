package main

import (
	"time"
	"bytes"
	"encoding/gob"
	"github.com/labstack/gommon/log"
)

//定义区块
type Block struct {
	Timestamp int64		//时间线
	Data []byte			//交易数据
	PrevHash []byte		//上一个区块hash
	Hash []byte			//当前区块hash
	Nonce int			//随机数
}

//创建区块
func NewBlock(data string, prevHash []byte) *Block {
	//block是一个指针,取得一个对象初始化后的地址
	block := &Block{time.Now().Unix(), []byte(data),prevHash,[]byte{}, 0}
	pow := NewProofOfWork(block)	//创建一个工作量证明的挖矿对象
	nonce,hash := pow.Run()		//开始挖矿
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock("涛酱的创世区块", []byte{})
}

//对象转为二进制字节集,写入文件
func (block *Block) SerializeBlock() []byte {
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