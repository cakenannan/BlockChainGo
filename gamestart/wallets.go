package main

import (
	"log"
	"bytes"
	"encoding/gob"
	"crypto/elliptic"
	"io/ioutil"
	"os"
	"fmt"
)

type Wallets struct {
	Wallets map[string]*Wallet	//一个字符串对应一个钱包
}

//创建钱包集合,或抓取已经存在的钱包集合
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID)
	return &wallets, err
}

//创建一个钱包
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet	//保存到集合
	return address
}

//从文件中读取钱包集合
func (ws *Wallets) LoadFromFile(nodeID string) error {
	myWalletFile := fmt.Sprintf(walletFile, nodeID)
	if _,err := os.Stat(myWalletFile); os.IsNotExist(err) {//判断文件是否存在
		return err
	}
	fileContent, err := ioutil.ReadFile(myWalletFile)	//读取文件
	if err != nil {
		log.Panic(err)
	}
	//读取文件二进制并解析
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	ws.Wallets = wallets.Wallets
	return nil
}

//钱包集合保存到文件
func (ws Wallets) SaveToFile(nodeID string) {
	myWalletFile := fmt.Sprintf(walletFile, nodeID)
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(myWalletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

//抓取钱包地址集合
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

//根据地址抓取钱包
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}