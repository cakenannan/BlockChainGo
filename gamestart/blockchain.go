package main

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"encoding/hex"
	"os"
)

const dbFile = "blockchain.db"		//数据库文件名，在当前目录下
const blockBucket = "blocks"			//bucket名称
const genesisCoinbaseData = "涛酱的创世区块数据"

type BlockChain struct {
	tip []byte		//存储最后一个块的哈希		二进制数据
	db *bolt.DB		//数据库
}

type BlockChainIterator struct {
	currentHash []byte	//当前的hash
	db *bolt.DB			//数据库
}

//挖矿
func (chain *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte		//最后一个块的hash
	err := chain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash = bucket.Get([]byte("l"))	//取出链上最后一个hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newblock := NewBlock(transactions, lastHash)	//创建新区块
	err = chain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put(newblock.Hash, newblock.Serialize())//插入数据
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("l"),newblock.Hash)		//更新最后hash值
		if err != nil {
			log.Panic(err)
		}
		chain.tip = newblock.Hash
		return nil
	})
}

//查找未被花销的交易
func (chain *BlockChain)FindUnspendTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOS := make(map[string][]int)
	bci := chain.Iterator()
	for {
		block := bci.Next()
		for _,tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)	//获取交易id
			Outputs:
				for outindex, out := range tx.Vout {
					if spentTXOS[txID] != nil {
						for _, spentOut := range spentTXOS[txID] {
							if spentOut == outindex {
								continue Outputs	//循环到不等为止
							}
						}
					}
					if out.CanBeUnlockWith(address) {
						unspentTXs = append(unspentTXs, *tx)
					}
				}
			if !tx.IsCoinBase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {	//是否可以锁定
						inTXID := hex.EncodeToString(in.Txid)
						spentTXOS[inTXID] = append(spentTXOS[inTXID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {	//创世区块
			break
		}
	}
	
	return unspentTXs
}

//获取所有未被花销的交易
func (chain *BlockChain)FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := chain.FindUnspendTransactions(address)//查找所有
	for _, tx := range unspentTransactions {
		for _,out := range tx.Vout {
			if out.CanBeUnlockWith(address) {	//判断是否锁定
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

//查找可以转账的交易
func (chain *BlockChain)FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := chain.FindUnspendTransactions(address)
	accmulated := 0		//累计
	Work:
		for _, tx := range unspentTXs {
			txID := hex.EncodeToString(tx.ID)
			for outindex, out := range tx.Vout {
				if out.CanBeUnlockWith(address) && accmulated < amount {
					accmulated += out.Value	//统计金额
					unspentOutputs[txID] = append(unspentOutputs[txID], outindex)//序列叠加
					if accmulated >= amount {
						break Work
					}
				}
			}
		}

	return accmulated,unspentOutputs
}

//判断数据库是否存在
func dbExits() bool {
	if _,err := os.Stat(dbFile);os.IsNotExist(err) {
		return false
	}
	return true
}

//增加一个区块
//func (chain *BlockChain) AddBlock(transactions []*Transaction) {
//	var lastHash []byte
//	err := chain.db.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte(blockBucket)) //取得bucket
//		lastHash = bucket.Get([]byte("l"))       //取得最后一块的hash
//		return nil
//	})
//	if err != nil {
//		log.Panic(err)
//	}
//	newblock := NewBlock(transactions, lastHash) //创建一个新的区块
//	err = chain.db.Update(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte(blockBucket))
//		err := bucket.Put(newblock.Hash, newblock.Serialize()) //压入块数据
//		if err != nil {
//			log.Panic(err)
//		}
//		err = bucket.Put([]byte("l"), newblock.Hash)	//刷新最后一个区块hash数据, l -> 链中最后一个块的 hash
//		if err != nil {
//			log.Panic(err)
//		}
//		chain.tip = newblock.Hash
//		return nil
//	})
//	if err != nil {
//		log.Panic(err)
//	}
//}

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
func NewBlockChain(address string) *BlockChain {
	if !dbExits() {
		fmt.Println("数据库不存在,创建")
		os.Exit(1)
	}
	var tip []byte
	db,err := bolt.Open(dbFile, 0600, nil)		//打开数据库
	if err != nil {
		log.Panic(err)
	}
	//处理数据更新
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))	//按照名称打开数据库的bucket,相当于一个表格或者说类别
		// TODO
		//if bucket == nil {
		//	fmt.Println("当前区块链没有数据库,创建一个新的")
		//	genesis := NewGenesisBlock()	//创建创世区块
		//	bucket,err := tx.CreateBucket([]byte(blockBucket))//创建一个数据库bucket
		//	if err != nil {
		//		log.Panic(err)
		//	}
		//	err = bucket.Put(genesis.Hash, genesis.Serialize()) //存入区块数据,hash为键,块对象序列化数据为值
		//	if err != nil {
		//		log.Panic(err)
		//	}
		//	err = bucket.Put([]byte("l"), genesis.Hash)	//存入数据,l -> 链中最后一个块的 hash
		//	if err != nil {
		//		log.Panic(err)
		//	}
		//	tip = genesis.Hash	//取得hash
		//} else {
		//	tip = bucket.Get([]byte("l"))
		//}

		tip = bucket.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc:= BlockChain{tip, db}	//创建一个区块链
	return &bc
}

func createBlockChain(address string) *BlockChain {
	if dbExits() {
		fmt.Println("数据库存在")
	}
	var tip []byte
	db,err := bolt.Open(dbFile, 0600, nil)		//打开数据库
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinBaseTX(address, genesisCoinbaseData)	//创世区块交易
		genesis := NewGenesisBlock(cbtx)	//创世区块
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})

	bc:= BlockChain{tip, db}	//创建一个区块链
	return &bc
}