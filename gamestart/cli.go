package main

import (
	"fmt"
	"os"
	"strconv"
	"flag"
	"github.com/labstack/gommon/log"
)

//命令行接口
type CLI struct {
	chain *BlockChain
}

//用法
func (cli *CLI) PrintUsage()  {
	fmt.Println("用法如下")
	fmt.Println("addblock 向区块链增加块")
	fmt.Println("showchain 显示区块链")
}

func (cli *CLI) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()	//显示用法
		os.Exit(1)
	}
}

func (cli *CLI) AddBlock(data string) {
	cli.chain.AddBlock(data)	//增加区块
	fmt.Println("区块增加成功")
}

func (cli *CLI) ShowBlockChain() {
	bci := cli.chain.Iterator()	//创建迭代器
	for {
		block := bci.Next()	//取得下一个区块
		fmt.Printf("上一区块hash%x\n", block.PrevHash)
		fmt.Printf("数据:%s\n", block.Data)
		fmt.Printf("当前区块hash:%x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s", strconv.FormatBool(pow.Validate()))
		fmt.Println("\n")

		if len(block.PrevHash) == 0 {	//循环到创世区块,终止
			break
		}
	}
}

//入口
func (cli *CLI) Run() {
	cli.ValidateArgs()	//校验
	//处理命令行参数
	addblockcmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)

	addBlockData := addblockcmd.String("data","","Block Data")
	switch os.Args[1] {
	case "addblock":
		err := addblockcmd.Parse(os.Args[2:])	//解析参数
		if err != nil {
			log.Panic(err)
		}
	case "showchain":
		err := showchaincmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUsage()
		os.Exit(1)
	}
	if addblockcmd.Parsed() {
		if *addBlockData == "" {
			addblockcmd.Usage()
			os.Exit(1)
		} else {
			cli.AddBlock(*addBlockData)	//增加区块
		}
	}
	if showchaincmd.Parsed() {
		cli.ShowBlockChain()	//显示区块链
	}
}
