package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

const version = byte(0x00)		//钱包版本
const walletFile = "wallet.dat"	//钱包文件
const addressChecksumlen = 4		//检测地址长度

type Wallet struct {
	PrivateKey ecdsa.PrivateKey	//私钥 钱包的权限文件
	PublicKey []byte	//公钥 收款地址
}

//创建钱包
func NewWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

//创建密钥对
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()	//创建加密算法
	private,err := ecdsa.GenerateKey(curve, rand.Reader)	//生成key
	if err != nil {
		log.Panic(err)
	}
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, public
}

//对公钥hash
func HashPubKey(pubkey []byte) []byte {
	publicsha256 := sha256.Sum256(pubkey)
	R160Hash := ripemd160.New()
	_,err := R160Hash.Write(publicsha256[:])
	if err != nil {
		log.Panic(err)
	}
	pubR160Hash := R160Hash.Sum(nil)
	return pubR160Hash
}

//抓去钱包地址
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)
	fullPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullPayload)
	return address
}

//校验钱包地址
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Encode([]byte(address))
	actualchecksum := pubKeyHash[len(pubKeyHash)-addressChecksumlen:]
	version := pubKeyHash[0] //取得版本
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-addressChecksumlen]
	targetCheckSum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualchecksum,targetCheckSum) == 0
}

//检验公钥
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumlen]
}