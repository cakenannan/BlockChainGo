package main

import (
	"time"
	"crypto/sha256"
	"encoding/hex"
	"github.com/labstack/echo"
	"fmt"
)

//区块模型
type BlockModel struct {
	Id int64	//索引
	TimeStamp string	//区块创建时间标识
	BPM int	//每分钟心跳频率
	Hash string		//sha256实现的区块哈希
	PreHash string	//上一个区块的哈希
}

//区块链,数组
var BlockChain = make([]BlockModel, 0)

//创建创世区块
func init() {
	block := BlockModel{0,time.Now().String(),0,"",""}	//创建区块
	BlockChain = append(BlockChain,	block)	//添加到链

}

//哈希处理
func calcHash(block BlockModel) string {
	record := string(block.Id) + block.TimeStamp + string(block.BPM) + block.PreHash
	myhash := sha256.New()	//创建算法对象
	myhash.Write([]byte(record))	//加入数据
	resultHash := myhash.Sum(nil)	//计算出hash值
	return hex.EncodeToString(resultHash)	//编码为字符串
}

//区块验证
func isBlockValid(newBlock, lastBlock BlockModel) bool {
	//id要顺序
	if lastBlock.Id + 1 != newBlock.Id {
		return false
	}
	//哈希校验
	if lastBlock.Hash != newBlock.PreHash {
		return false
	}
	//验证数据未被修改
	if calcHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

//创建区块
func createBlock(ctx echo.Context) error {
	//处理心跳信息
	type message struct {
		BPM int
	}
	var myMessage = message{}
	if err := ctx.Bind(&myMessage); err != nil {	//绑定消息处理
		panic(err)
	}
	lastblock := BlockChain[len(BlockChain) - 1]	//上一个区块
	//创建新区块
	newblock := BlockModel{}
	newblock.Id = lastblock.Id + 1
	newblock.TimeStamp = time.Now().String()
	newblock.BPM = myMessage.BPM
	newblock.PreHash = lastblock.Hash
	newblock.Hash = calcHash(newblock)
	if isBlockValid(newblock, lastblock) {
		BlockChain = append(BlockChain, newblock)
		fmt.Println("创建区块成功",BlockChain[len(BlockChain) - 1].Id)
	} else {
		fmt.Println("创建区块失败")
	}
	return ctx.JSON(200, newblock)
}

func main() {
	echoserver := echo.New()	//创建服务器
	echoserver.GET("/", func(context echo.Context) error {
		return context.JSON(200, BlockChain)
	})
	echoserver.POST("/", createBlock)
	echoserver.Logger.Fatal(echoserver.Start(":8848"))
}
