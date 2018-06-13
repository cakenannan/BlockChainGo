package main

import (
	"math"
	"math/big"
	"bytes"
	"fmt"
	"crypto/sha256"
)

var (
	maxNonce = math.MaxInt64	//最大的64位整数
)

// 在比特币中，当一个块被挖出来以后，“target bits” 代表了区块头里存储的难度，也就是开头有多少个 0。
// 这里的 16 指的是算出来的哈希前 16 位必须是 0，如果用 16 进制表示，就是前 6 位必须是 0
const targetBits = 16	//对比的位数	挖矿难度值

type ProofOfWork struct {
	block *Block	//区块
	target *big.Int	//存储计算哈希对比的特定整数
}

//创建一个工作量证明的挖矿对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)		//初始化目标整数
	// 我们将 big.Int 初始化为 1，然后左移 256 - targetBits 位。256 是一个 SHA-256 哈希的位数，我们将要使用的是 SHA-256 哈希算法。
	// target（目标） 的 16 进制形式为：0x10000000000000000000000000000000000000000000000000000000000
	target.Lsh(target, uint(256 - targetBits))		//数据转换
	pow := &ProofOfWork{block, target}	//创建对象
	return pow
}

//准备数据进行挖矿计算
func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevHash,
			pow.block.HashTransactions(),	//交易hash
			IntToHex(pow.block.Timestamp),	//时间,转化为十六进制
			IntToHex(int64(targetBits)),	//位数
			IntToHex(int64(nonce)),			//保存工作量的nonce
		},[]byte{},
	)
	return data
}

//挖矿
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	var nonce = 0
	for nonce < maxNonce {				// maxNounce被设置成math.MaxInt64，防止溢出
		data := pow.PrepareData(nonce)	//准备数据
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x",hash)	//打印hash
		hashInt.SetBytes(hash[:])		// 将hash转为大整数
		if hashInt.Cmp(pow.target) == -1 {	//挖矿校验	大小比较,小于目标值则break
			break
		} else {
			nonce++
		}
	}
	fmt.Println("\n\n")
	return nonce, hash[:]
}

//校验
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := (hashInt.Cmp(pow.target) == -1)	//检验
	return isValid
}