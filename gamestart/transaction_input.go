package main

import "bytes"

//输入
type TXInput struct {
	Txid []byte			//一个输入引用了之前交易的一个输出,之前交易id
	Vout int			//该输出在那笔交易中所有输出的索引
	Signature []byte	//签名
	PubKey []byte		//公钥
}

//key检测地址和交易
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
