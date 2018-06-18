package main

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"io"
	"encoding/hex"
	"io/ioutil"
)

const protocol = "tcp"	//网络协议
const nodeVersion = 1		//版本
const commandlength = 12	//命令行长度	//约定消息前12个字节为命令名,后面的字节为数据

var nodeAddress string                     //当前节点地址
var miningAddress string                   //挖矿节点地址
var knowNodes = []string{"localhost:3000"} //已知节点		//这里对中心节点硬编码
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)	//内存池

type addr struct {
	Addrlist []string	//节点地址集合
}

type block struct {	//块消息
	AddrFrom string
	Block []byte		//块的序列化数据
}

type getblocks struct {	//请求块 消息
	AddrFrom string
}

type tx struct {
	AddrFrom string
	Transaction []byte	//交易的序列化数据
}

type getdata struct {		//用于某个块或交易的请求
	AddrFrom string
	Type string
	ID []byte				//块或者交易id
}

//比特币使用 inv 来向其他节点展示当前节点有什么块和交易。
// 它没有包含完整的区块链和交易，仅仅是哈希而已。
// Type 字段表明了这是块还是交易。
type inv struct {
	AddrFrom string
	Type string
	Items [][]byte
}

type verzion struct {
	Version int			//版本号
	BestHeight int		//区块链节点高度
	AddrFrom string		//发送者的地址
}

//命令转化为字节
func commandToBytes(command string) []byte {
	var bytes [commandlength]byte
	for i, char := range command {
		bytes[i] = byte(char)
	}
	return bytes[:]
}

//字节转化为命令
func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

//提取命令
func extractCommand(request []byte) []byte {
	return request[:commandlength]
}

//请求块
func requestBlocks() {
	for _, node := range knowNodes {	//给所有已知节点发送请求
		sendGetBlocks(node)
	}
}

//发送数据
func sendData(addr string, data []byte) {
	conn,err := net.Dial(protocol, addr)	//建立网络链接
	if err != nil {
		fmt.Printf("%s 地址不可访问\n", addr)
		//去掉不可访问的地址
		var updataNodes []string
		for _, node := range knowNodes {
			if node != addr {
				updataNodes = append(updataNodes, node)
			}
		}
		knowNodes = updataNodes
	}
	defer conn.Close()
	_,err = io.Copy(conn, bytes.NewReader(data))	//发送数据
	if err != nil {
		log.Panic(err)
	}
}

//发送请求块消息
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}

//发送请求的数据
func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)
}

//发送地址
func sendAddr(address string) {
	nodes := addr{knowNodes}
	nodes.Addrlist = append(nodes.Addrlist, nodeAddress)//在已知节点中追加当前节点
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)
	sendData(address, request)		//发送数据
}

//发送块
func sendBlock(address string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(address, request)
}

//发送库存数据
func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}	//库存数据
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)
	sendData(address, request)
}

//发送交易
func sendTx(address string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"),payload...)
	sendData(address, request)
}

//发送版本信息
func sendVersion(address string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	fmt.Println("sendVersion addfrom=", nodeAddress)
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"),payload...)
	sendData(address, request)
}

//处理地址
func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr
	buff.Write(request[commandlength:])	//取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	knowNodes = append(knowNodes, payload.Addrlist...)	//追加
	fmt.Printf("已经有了%d个已知节点\n", len(knowNodes))
	requestBlocks()
}

//处理块消息
func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	block := DeserializeBlock(blockData)
	fmt.Printf("收到一个新的区块,hash=%x\n",block.Hash)
	bc.AddBlock(block)
	fmt.Printf("增加一个区块hash=%x\n",block.Hash)
	//当接收到一个新块时，我们把它放到区块链里面。
	// 如果还有更多的区块需要下载，我们继续从上一个下载的块的那个节点继续请求。
	// 当最后把所有块都下载完后，对 UTXO 集进行重新索引。
	if len(blocksInTransit) > 0 {
		blockhash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockhash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
		//UTXOSet.Update(block)
	}
}

//处理
func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("收到inv %d %s\n",len(payload.Items), payload.Type)
	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockhash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockhash)
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockhash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}
	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

//处理请求块消息的请求,返回hash值列表
func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

//处理请求数据消息的请求
func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block" {	//如果请求块,发送块回去
		fmt.Printf("handleGetData type=block,id=%x\n",payload.ID)
		fmt.Println(payload.ID)
		block,err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)		//todo 这里发送的是块数据
	}
	if payload.Type == "tx" {		//如果请求交易,发送交易回去
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	}
}

//处理交易
func handleTransaction(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload tx
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	txData := payload.Transaction	//交易数据
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx
	//fmt.Println(nodeAddress, knowNodes[0])
	if nodeAddress == knowNodes[0] {			//如果是中心节点,分发交易
		for _, node := range knowNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {	//挖矿节点挖矿打包交易
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction
			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}
			if len(txs) == 0 {
				fmt.Println("没有任何交易,等待新的交易加入")
				return
			}
			cbtx := NewCoinBaseTX(miningAddress, "")
			txs = append(txs, cbtx)
			newblock := bc.MineBlock(txs)
			utxoSet := UTXOSet{bc}
			utxoSet.Reindex()
			//utxoSet.Update(newblock)
			fmt.Printf("新的区块已经挖掘到\n")
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)	//交易打包到新的块中,在内存池中删除该交易
			}
			for _, node := range knowNodes {	//挖矿成功后广播
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newblock.Hash})
				}
			}
			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload verzion
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	fmt.Println("handleversion addrfrom=" + payload.AddrFrom)
	if myBestHeight < foreignerBestHeight {	//比较区块链长度
		sendGetBlocks(payload.AddrFrom)		//发送请求块消息
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)	//发送版本消息
	}
	if !nodeIsKnow(payload.AddrFrom) {
		knowNodes = append(knowNodes, payload.AddrFrom)
	}
}

//处理网络链接
func handleConnection(conn net.Conn, bc *BlockChain) {
	request,err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandlength])	//取出命令
	fmt.Printf("收到command%s\n", command)
	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "tx":
		handleTransaction(request, bc)
	case "version":
		handleVersion(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	default:
		fmt.Println("未知命令")
	}
	conn.Close()
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s",nodeID)
	miningAddress = minerAddress
	ln,err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	bc := NewBlockChain(nodeID)
	if nodeAddress != knowNodes[0] {
		//如果当前节点不是中心节点,必须向中心节点发送version消息来查询自己的区块链是否已过时
		sendVersion(knowNodes[0], bc)
	}
	for {
		conn,err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)	//异步处理
	}
}

//判断是否是已知节点
func nodeIsKnow(addr string) bool {
	for _, node := range knowNodes {
		if node == addr {
			return true
		}
	}
	return false
}

//将编码逻辑封装
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
