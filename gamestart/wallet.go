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

const version = byte(0x00)      //钱包版本
const walletFile = "wallet_%s.dat" //钱包文件
const addressChecksumLen = 4    //公钥hash的摘要位数

type Wallet struct {
	PrivateKey ecdsa.PrivateKey	//私钥
	PublicKey []byte			//公钥
}

//创建钱包
func NewWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

//创建密钥对
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()	//曲线
	private,err := ecdsa.GenerateKey(curve, rand.Reader)	//私钥
	if err != nil {
		log.Panic(err)
	}
	// 公钥
	// 在基于椭圆曲线的算法中，公钥是曲线上的点。因此，公钥是 X，Y 坐标的组合。
	// 在比特币中，这些坐标会被连接起来，然后形成一个公钥。
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, public
}

//通过钱包公钥生成address 分三部分,版本号+公钥hash+公钥hash的四位摘要
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)							//对公钥进行hash
	versionedPayload := append([]byte{version}, pubKeyHash...)		//给哈希加上版本号作为前缀
	checksum := checksum(versionedPayload)							//生成hash的四位摘要
	fullPayload := append(versionedPayload, checksum...)			//连接上四位摘要
	address := Base58Encode(fullPayload)							//进行Base58编码
	return address
}

//校验钱包地址
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))                       	//解码
	actualchecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:] //取得公钥hash的四位摘要
	version := pubKeyHash[0]                                          	//取得版本
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-addressChecksumLen]     //取得公钥hash
	targetCheckSum := checksum(append([]byte{version}, pubKeyHash...))	//计算公钥hash的四位摘要
	return bytes.Compare(actualchecksum,targetCheckSum) == 0			//对比校验
}

//对公钥hash
func HashPubKey(pubkey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubkey) //先sha256
	R160Hasher := ripemd160.New()         //再RIPEMD160
	_,err := R160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	pubR160Hash := R160Hasher.Sum(nil)
	return pubR160Hash
}

//生成指定位数摘要
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}
