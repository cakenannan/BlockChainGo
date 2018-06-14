package main

import "bytes"

//输出
type TXOutput struct {
	Value int			//一定量的比特币
	PubKeyHash []byte	//
}

//锁住输出
func (out *TXOutput) Lock(address []byte) {
	pubkeyhash := Base58Encode(address)
	pubkeyhash = pubkeyhash[1:len(pubkeyhash)-4]	//截取有效hash
	out.PubKeyHash = pubkeyhash			//锁住,无法被修改
}

//是否被key锁住
func (out *TXOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubkeyHash) == 0
}

//创建一个输出
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}