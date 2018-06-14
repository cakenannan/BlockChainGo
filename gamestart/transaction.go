package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"
	"strings"
	"crypto/elliptic"
	"math/big"
)

const subsidy = 1000		//奖励,矿工挖矿给与的奖励

//交易,编号,输入,输出
type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

//序列化
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

//反序列化
func Deserialize(data []byte) Transaction {
	var transaction Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

//对象字符串展示方法重写
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction %x", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("input %d", i))
		lines = append(lines, fmt.Sprintf("TXID %x", input.Txid))
		lines = append(lines, fmt.Sprintf("OUT %d", input.Vout))
		lines = append(lines, fmt.Sprintf("Signature %d", input.Signature))
		lines = append(lines, fmt.Sprintf("PubKey %d", input.PubKey))
	}
	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("out %d", i))
		lines = append(lines, fmt.Sprintf("value %d", output.Value))
		lines = append(lines, fmt.Sprintf("OUT %x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}

//对交易事务进行hash
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())	//取得二进制进行hash计算
	return hash[:]
}

//签名
func (tx *Transaction)Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("以前的交易不正确")
		}
	}
	txCopy := tx.TrimmedCopy()	//拷贝副本
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil
		r,s,err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
}

//签名验证
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("之前的交易不正确")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash	//设置公钥

		r := big.Int{}
		s := big.Int{}
		siglen := len(vin.Signature)	//统计签名的长度
		r.SetBytes(vin.Signature[:(siglen/2)])
		s.SetBytes(vin.Signature[(siglen/2):])

		x := big.Int{}
		y := big.Int{}
		keylen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keylen/2)])
		y.SetBytes(vin.PubKey[(keylen/2):])

		dataToVerify := fmt.Sprintf("%x\n",txCopy)
		rawPubKey := ecdsa.PublicKey{curve,&x, &y}
		if !ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) {
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}
	return true
}

//用于签名的交易事务裁剪的副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

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

//挖矿交易
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		//data = fmt.Sprintf("挖矿奖励给%s",to)
		randData := make([]byte, 20)
		_,err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x",randData)
	}
	//输入
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}	//Vout挖矿所得为-1
	//输出
	txout := NewTXOutput(subsidy, to)
	//交易
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.SetID()
	return &tx
}

//转账交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	pubkeyhash := HashPubKey(wallet.PublicKey)

	//找到未花费outputs
	acc,validOutputs := bc.FindSpendableOutputs(pubkeyhash, amount)
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
			input := TXInput{txID, out, nil, wallet.PublicKey}	//输入的交易
			inputs = append(inputs, input)	//输出的交易
		}
	}
	// 对于outputs,包括数量amount指向to,和数量acc-amount指向from
	outputs = append(outputs, *NewTXOutput(amount, to))	//输出到to
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))	//找零到from
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}
