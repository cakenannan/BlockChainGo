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

//查找包含未被花费输出的交易
func (chain *BlockChain)FindUnspendTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOS := make(map[string][]int)
	bci := chain.Iterator()
	for {
		block := bci.Next()
		for _,tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				//检查out是否已经被花费
				if spentTXOS[txID] != nil {
					for _, spentOut := range spentTXOS[txID] {
						if spentOut == outIdx {
							continue Outputs	//循环到不等为止
						}
					}
				}
				//如果一个输出被一个地址锁定，并且这个地址恰好是我们要找的地址，那么这个输出就是我们想要的
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			//将给定地址所有能够解锁输出的输入聚集起来（这并不适用于 coinbase 交易，因为它们不解锁输出）
			if !tx.IsCoinBase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTXID := hex.EncodeToString(in.Txid)
						spentTXOS[inTXID] = append(spentTXOS[inTXID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {	//循环到创世区块
			break
		}
	}
	
	return unspentTXs
}

//获取所有未被花费的输出
func (chain *BlockChain)FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := chain.FindUnspendTransactions(address)//查找所有
	for _, tx := range unspentTransactions {
		for _,out := range tx.Vout {
			if out.CanBeUnlockedWith(address) { //判断是否解锁
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// 对所有的未花费交易进行迭代，并对它的值进行累加。
// 当累加值大于或等于我们想要传送的值时，它就会停止并返回累加值，
// 同时返回的还有通过交易 ID 进行分组的输出索引。
// 我们只需取出足够支付的钱就够了。
func (chain *BlockChain)FindSpendableOutputs(address string, amount int) (int, map[string][]int) {//key是string,value是int数组
	unspentOutputs := make(map[string][]int)
	unspentTXs := chain.FindUnspendTransactions(address)
	accmulated := 0		//累计
	Work:
		for _, tx := range unspentTXs {
			txID := hex.EncodeToString(tx.ID)
			for outIdx, out := range tx.Vout {
				if out.CanBeUnlockedWith(address) && accmulated < amount {
					accmulated += out.Value	//累加金额
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
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
		tip = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc:= BlockChain{tip, db}	//创建一个区块链
	return &bc
}

//创建一个区块链
func createBlockChain(address string) *BlockChain {
	if dbExits() {
		fmt.Println("数据库存在")
		os.Exit(1)
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