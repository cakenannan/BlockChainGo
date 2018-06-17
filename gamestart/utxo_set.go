package main

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
)

const utxoBucket = "chainstate"	//存储UTXO集的bucket

//二次封装区块链
type UTXOSet struct {
	blockchain *BlockChain
}

//查找并返回可使用尽可能足够的输出
func (utxo UTXOSet) FindSpendableOutputs(pubkeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := utxo.blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()
		for key, value := cur.First(); key != nil; key, value = cur.Next() {
			txID := hex.EncodeToString(key)
			outs := DeserializeOutputs(value)
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyhash) && accumulated < amount {	//可解锁并且数量还不够
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

//查找UTXO
func (utxoset UTXOSet) FindUTXO(pubkeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := utxoset.blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()
		for key, value := cur.First(); key != nil; key, value = cur.Next() {
			outs := DeserializeOutputs(value)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyhash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

//统计包含UTXO的交易数量
func (utxoset UTXOSet) CountTransactions() int {
	db := utxoset.blockchain.db
	counter := 0
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()
		for k, _ := cur.First(); k != nil; k,_ = cur.Next() {
			counter++
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return counter
}

//重建UTXO集,遍历区块链重建
func (utxo UTXOSet) Reindex() {
	db := utxo.blockchain.db
	bucketName := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)					//删除
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_,err = tx.CreateBucket(bucketName)					//再创建
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	UTXO := utxo.blockchain.FindUTXO()			//遍历块找到UTXO集
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		for txID, outs := range UTXO {
			txid,err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(txid, outs.Serialize())	//加到数据库中
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//更新 UTXO 集,当生成新的区块的时候
func (utxo UTXOSet) Update(block *Block) {
	db := utxo.blockchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if !tx.IsCoinBase() {
				for _, vin := range tx.Vin {	//循环输入,找到引用的输出,然后更新状态
					updateOuts := TXOutputs{}
					outsBytes := bucket.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {	//输入所引用的输出未被花费,添加到集合中
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}
					if len(updateOuts.Outputs) == 0 { //如果对应交易id的所有输出都被花费,移除该交易
						err := bucket.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := bucket.Put(vin.Txid, updateOuts.Serialize())//更新对应交易id的输出集,移除掉已花费的
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}
			newOutputs := TXOutputs{}			//处理交易中新的输出
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := bucket.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
