package main

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
)

const dbFile = "blockchain.db"		//数据库文件名，在当前目录下
const blockBucket = "blocks"			//名称，

type BlockChain struct {
	tip []byte		//存储最后一个块的哈希		二进制数据
	db *bolt.DB		//数据库
}

type BlockChainIterator struct {
	currentHash []byte	//当前的hash
	db *bolt.DB			//数据库
}

//增加一个区块
func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := chain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取得bucket
		lastHash = bucket.Get([]byte("l"))       //取得最后一块的hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newblock := NewBlock(data, lastHash) //创建一个新的区块
	err = chain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put(newblock.Hash, newblock.Serialize()) //压入块数据
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("l"), newblock.Hash)	//刷新最后一个区块hash数据, l -> 链中最后一个块的 hash
		if err != nil {
			log.Panic(err)
		}
		chain.tip = newblock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//根据区块链创建迭代器
func (chain *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{chain.tip, chain.db}
	return bci
}

//取得下一个区块
func (it *BlockChainIterator) Next() *Block {
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		encodedBlock := bucket.Get(it.currentHash)	//抓取二进制数据
		block = DeserializeBlock(encodedBlock)	//解码
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	it.currentHash = block.PrevHash
	return block
}

//新建一个区块链
func NewBlockChain() *BlockChain {
	var tip []byte
	db,err := bolt.Open(dbFile, 0600, nil)		//打开数据库
	if err != nil {
		log.Panic(err)
	}
	//处理数据更新
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))	//按照名称打开数据库的bucket,相当于一个表格或者说类别
		if bucket == nil {
			fmt.Println("当前区块链没有数据库,创建一个新的")
			genesis := NewGenesisBlock()	//创建创世区块
			bucket,err := tx.CreateBucket([]byte(blockBucket))//创建一个数据库bucket
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(genesis.Hash, genesis.Serialize()) //存入区块数据,hash为键,块对象序列化数据为值
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put([]byte("l"), genesis.Hash)	//存入数据,l -> 链中最后一个块的 hash
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash	//取得hash
		} else {
			tip = bucket.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc:= BlockChain{tip, db}	//创建一个区块链
	return &bc
}