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
func DeserializeTransaction(data []byte) Transaction {
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

//对交易序列化,并进行hash,得到要签名的数据
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())	//取得二进制进行hash计算
	return hash[:]
}

//考虑到交易解锁的是之前的输出，然后重新分配里面的价值，并锁定新的输出，那么必须要签名以下数据：
//	1,存储在已解锁输出的公钥哈希。它识别了一笔交易的“发送方”。
//	2,存储在新的锁定输出里面的公钥哈希。它识别了一笔交易的“接收方”。
//	3,新的输出值。
//在比特币中，锁定/解锁逻辑被存储在脚本中，它们被分别存储在输入和输出的 ScriptSig 和 ScriptPubKey 字段。
//由于比特币允许这样不同类型的脚本，它对 ScriptPubKey 的整个内容进行了签名。
//签名	并不是对交易签名,而是对去除部分内容的输入副本签名,输入里面存储了被引用输出的 ScriptPubKey
func (tx *Transaction)Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("之前的交易不正确")
		}
	}
	txCopy := tx.TrimmedCopy()				//交易副本
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]	//根据id找到输入引用的输出的那笔交易
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash	//引用输出的公钥哈希赋值给输入
		txCopy.ID = txCopy.Hash()			//Hash得到要签名的数据,给id赋值
		txCopy.Vin[inID].PubKey = nil
		r,s,err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)	//对txCopy.ID签名
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
		txCopy.ID = txCopy.Hash()		//求得ID,即签名的数据
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		siglen := len(vin.Signature)	//签名的长度
		r.SetBytes(vin.Signature[:(siglen/2)])
		s.SetBytes(vin.Signature[(siglen/2):])

		x := big.Int{}
		y := big.Int{}
		keylen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keylen/2)])
		y.SetBytes(vin.PubKey[(keylen/2):])

		rawPubKey := ecdsa.PublicKey{curve,&x, &y}	//获取公钥
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {		//验证 公钥,数据,签名
			return false
		}
	}
	return true
}

//用于签名的交易事务裁剪的副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		//输入的Signature和PubKey被置为nil
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

//挖矿交易
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_,err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}
	//输入
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}	//Vout挖矿所得为-1
	//输出
	txout := NewTXOutput(subsidy, to)
	//交易
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()
	return &tx
}

//转账交易
func NewUTXOTransaction(wallet *Wallet, to string, amount int, utxoset *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubkeyhash := HashPubKey(wallet.PublicKey)

	//找到未花费outputs
	acc,validOutputs := utxoset.FindSpendableOutputs(pubkeyhash, amount)
	//fmt.Printf("acc=%d,amount=%d\n",acc,amount)
	if acc < amount {
		log.Panic("交易金额不足")
	}
	// 对于每个找到的输出,创建一个引用该输出的输入
	for txid,outs := range validOutputs {				//先循环得到交易id和该交易中对应输出的索引集合
		txID,err := hex.DecodeString(txid)	//转为字节数组
		if err != nil {
			log.Panic(err)
		}
		for _,out := range outs {						//循环同一交易下多个输出
			input := TXInput{txID, out, nil, wallet.PublicKey}	//输入
			inputs = append(inputs, input)
		}
	}
	from := fmt.Sprintf("%s", wallet.GetAddress())
	// 对于outputs,包括数量amount指向to,和数量acc-amount指向from
	outputs = append(outputs, *NewTXOutput(amount, to))	//输出到to
	//因为输出不可分,所以acc可能大于amount,要找零
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))	//找零到from
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	utxoset.blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}
