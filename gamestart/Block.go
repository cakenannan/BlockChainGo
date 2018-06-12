package main

import (
	"strconv"
	"bytes"
	"crypto/sha256"
	"time"
)

//定义区块
type Block struct {
	Timestamp int64		//时间线
	Data []byte			//交易数据
	PrevHash []byte		//上一个区块hash
	Hash []byte			//当前区块hash
}

//设定结构体对象hash
func (block *Block) SetHash() {
	//处理当前时间,转化为10进制字符串,再转化为字节集合
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))
	//叠加要hash的数据
	headers := bytes.Join([][]byte{block.PrevHash, block.Data, timestamp},[]byte{})
	//计算出hash地址
	hash := sha256.Sum256(headers)
	//设置hash
	block.Hash = hash[:]
}

//创建区块
func NewBlock(data string, prevHash []byte) *Block {
	//block是一个指针,取得一个对象初始化后的地址
	block := &Block{time.Now().Unix(), []byte(data),prevHash,[]byte{}}
	block.SetHash()		//设置当前hash
	return block
}

//创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock("涛酱的创世区块", []byte{})
}
