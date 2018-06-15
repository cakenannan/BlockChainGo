package main

import "bytes"

type TXInput struct {
	Txid []byte			//一个输入引用了之前交易的一个输出,之前交易id
	Vout int			//引用的之前交易输出在之前那笔交易中所有输出的索引
	Signature []byte	//签名
	PubKey []byte		//公钥
}

//校验,检查输入使用了指定密钥来解锁一个输出
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
