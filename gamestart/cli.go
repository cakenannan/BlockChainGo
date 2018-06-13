package main

import (
	"fmt"
	"os"
	"strconv"
	"flag"
	"log"
)

//命令行接口
type CLI struct {
}

//用法
func (cli *CLI) PrintUsage()  {
	fmt.Println("用法如下")
	fmt.Println("send -from From -to To -amount Amount 转账")
	fmt.Println("showchain 显示区块链")
	fmt.Println("createblockchain -address 地址 根据地址创建区块链")
	fmt.Println("getbalance -address 地址 根据地址查询金额")
}

func (cli *CLI) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()	//显示用法
		os.Exit(1)
	}
}

func (cli *CLI) ShowBlockChain() {
	bc := NewBlockChain("")
	defer bc.db.Close()
	bci := bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("上一区块hash%x\n", block.PrevHash)
		fmt.Printf("当前区块hash:%x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s \n\n",strconv.FormatBool(pow.Validate()))
		if len(block.PrevHash) == 0 {	//循环到创世区块,终止
			break
		}
	}
}

func (cli *CLI) Send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)	//转账
	bc.MineBlock([]*Transaction{tx})	//挖矿确认交易,记账成功
	fmt.Println("交易成功")
}

func (cli *CLI) createBlockChain(address string) {
	bc := createBlockChain(address)		//创建区块链
	bc.db.Close()
	fmt.Println("创建成功", address)
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.db.Close()
	balance := 0
	UTXOs := bc.FindUTXO(address)	//查找交易金额
	for _, out := range UTXOs {
		balance += out.Value	//取出金额
	}
	fmt.Printf("查询的地址为:%s,金额为:%d \n", address, balance)
}

//入口
func (cli *CLI) Run() {
	cli.ValidateArgs()	//校验
	//处理命令行参数
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)

	sendfrom := sendcmd.String("from","","from地址")
	sendto := sendcmd.String("to","","to地址")
	sendamount := sendcmd.Int("amount",0,"amount金额")
	getbalanceaddress := getbalancecmd.String("address","","查询余额地址")
	createblockchainaddress := createblockchaincmd.String("address","","创建区块链地址")

	switch os.Args[1] {
	case "showchain":
		err := showchaincmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendcmd.Parse(os.Args[2:])	//解析参数
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getbalancecmd.Parse(os.Args[2:])	//解析参数
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createblockchaincmd.Parse(os.Args[2:])	//解析参数
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUsage()
		os.Exit(1)
	}
	if showchaincmd.Parsed() {
		cli.ShowBlockChain()	//显示区块链
	}
	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		} else {
			cli.Send(*sendfrom, *sendto, *sendamount)
		}
	}
	if getbalancecmd.Parsed() {
		if *getbalanceaddress == "" {
			getbalancecmd.Usage()
			os.Exit(1)
		} else {
			cli.getBalance(*getbalanceaddress)
		}
	}
	if createblockchaincmd.Parsed() {
		if *createblockchainaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		} else {
			cli.createBlockChain(*createblockchainaddress)
		}
	}
}
