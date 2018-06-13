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

//输入
type TXInput struct {
	Txid []byte			//交易id
	Vout int			//保存交易中的一个output索引
	ScriptSig string	//保存了一个任意用户定义的钱包地址
}

//检查地址是否启动事务
func (input *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

//检查交易事务是否为coinbase,挖矿得来的奖励币
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//设置交易id
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

//输出
type TXOutput struct {
	Value int			//币值
	ScriptPubKey string	//
}

//是否可以解锁输出
func (out *TXOutput) CanBeUnlockWith(unlockingData string) bool {
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
	//输入奖励
	txin := TXInput{[]byte{}, -1, data}	//Vout挖矿所得为-1
	//输入奖励
	txout := TXOutput{subsidy, to}
	//交易
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	return &tx
}

//转账交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	acc,validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("交易金额不足")
	}
	for txid,outs := range validOutputs {	//循环遍历无效输出
		txID,err := hex.DecodeString(txid)	//解码
		if err != nil {
			log.Panic(err)
		}
		for _,out := range outs {
			input := TXInput{txID, out, from}	//输入的交易
			inputs = append(inputs, input)	//输出的交易
		}
	}
	//交易叠加
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
