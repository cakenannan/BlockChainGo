package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	Value int			//数量
	PubKeyHash []byte	//公钥hash
}

//创建一个输出,指向address
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

//锁住输出,对输出的公钥hash赋值,让输出有指向address
func (out *TXOutput) Lock(address []byte) {
	pubkeyHash := Base58Decode(address)			//解码
	pubkeyHash = pubkeyHash[1:len(pubkeyHash)-4] //截取有效hash
	out.PubKeyHash = pubkeyHash                  //对输出的公钥hash赋值
}

//校验,校验公钥hash,即校验地址
func (out *TXOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubkeyHash) == 0
}

type TXOutputs struct {
	Outputs []TXOutput		//要序列化，字段首字母必须大写，干干干！！！
}

func (outputs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(outputs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TXOutputs {
	var txoutputs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&txoutputs)
	if err != nil {
		log.Panic(err)
	}
	return txoutputs
}
