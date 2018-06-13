package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"encoding/hex"
)

const subsidy = 10		//奖励,矿工挖矿给与的奖励

//是否为coinbase交易
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//设置交易id,为交易hash值
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

//输入
type TXInput struct {
	Txid []byte			//一个输入引用了之前交易的一个输出,之前交易id
	Vout int			//该输出在那笔交易中所有输出的索引
	//脚本 提供了可解锁输出结构里面 ScriptPubKey 字段的数据
	//提供的数据是正确的，那么输出就会被解锁，然后被解锁的值就可以被用于产生新的输出；
	// 如果数据不正确，输出就无法被引用在输入中，或者说，无法使用这个输出
	ScriptSig string	//由于我们还没有实现地址，所以目前 ScriptSig 将仅仅存储一个用户自定义的任意钱包地址
}

//校验输入
func (input *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

//输出
type TXOutput struct {
	Value int			//一定量的比特币
	ScriptPubKey string	//锁定脚本,要花费这笔钱,必须要解锁该脚本
}

//检验输出
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

//交易,编号,输入,输出
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

//挖矿交易
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("挖矿奖励给%s",to)
	}
	//输入
	txin := TXInput{[]byte{}, -1, data}	//Vout挖矿所得为-1
	//输出
	txout := TXOutput{subsidy, to}
	//交易
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	return &tx
}

//转账交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	//找到未花费outputs
	acc,validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("交易金额不足")
	}
	// 对于每个找到的输出,创建一个引用该输出的输入
	for txid,outs := range validOutputs {
		txID,err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _,out := range outs {
			input := TXInput{txID, out, from}	//输入的交易
			inputs = append(inputs, input)	//输出的交易
		}
	}
	// 对于outputs,包括数量amount指向to,和数量acc-amount指向from
	outputs = append(outputs, TXOutput{amount, to})	//输出到to
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})	//找零到from
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
