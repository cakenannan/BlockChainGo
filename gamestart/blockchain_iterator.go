package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockChainIterator struct {
	currentHash []byte	//当前的hash
	db *bolt.DB			//数据库
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
