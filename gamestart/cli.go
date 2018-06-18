package main

import (
	"fmt"
	"os"
	"flag"
	"log"
)

//命令行接口
type CLI struct {}

//用法
func (cli *CLI) PrintUsage()  {
	fmt.Println("用法如下")
	fmt.Println("send -from From -to To -amount Amount -mine MINENOW 转账")
	fmt.Println("showchain 显示区块链")
	fmt.Println("createblockchain -address 地址 根据地址创建区块链")
	fmt.Println("getbalance -address 地址 根据地址查询金额")
	fmt.Println("createwallet 创建钱包")
	fmt.Println("listaddress 列出地址集合")
	fmt.Println("reindexutxo 重建索引")
	fmt.Println("startnode -miner ADDR 开启一个节点")
}

func (cli *CLI) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()	//显示用法
		os.Exit(1)
	}
}

//入口
func (cli *CLI) Run() {
	cli.ValidateArgs()	//校验

	//nodeID := os.Getenv("NODE_ID")
	//if nodeID == "" {
	//	fmt.Println("必须设置运行端口号")
	//	os.Exit(1)
	//}
	//nodeID  := "3000"		//中心节点
	//nodeID  := "3001"		//钱包节点
	nodeID  := "3002"		//挖矿节点


	//处理命令行参数
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listaddresscmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
	reindexutxocmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startnodecmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	sendfrom := sendcmd.String("from","","from地址")
	sendto := sendcmd.String("to","","to地址")
	sendamount := sendcmd.Int("amount",0,"amount金额")
	sendmine := sendcmd.Bool("mine",false,"是否立刻挖矿")
	getbalanceaddress := getbalancecmd.String("address","","查询余额地址")
	createblockchainaddress := createblockchaincmd.String("address","","创建区块链地址")
	startnodeminer := startnodecmd.String("miner","","miner地址")

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
	case "createwallet":
		err := createwalletcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddress":
		err := listaddresscmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexutxocmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startnodecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUsage()
		os.Exit(1)
	}

	if showchaincmd.Parsed() {
		cli.ShowBlockChain(nodeID)	//显示区块链
	}
	if createwalletcmd.Parsed() {
		cli.createWallet(nodeID)		//创建钱包
	}
	if listaddresscmd.Parsed() {
		cli.listAddress(nodeID)		//显示所有地址
	}
	if reindexutxocmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}
	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount, nodeID, *sendmine)
	}
	if getbalancecmd.Parsed() {
		if *getbalanceaddress == "" {
			getbalancecmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceaddress,nodeID)
	}
	if createblockchaincmd.Parsed() {
		if *createblockchainaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockchainaddress,nodeID)
	}
	if startnodecmd.Parsed() {
		if nodeID == "" {
			startnodecmd.Usage()
			os.Exit(1)
		}
		cli.startNode(*startnodeminer, nodeID)
	}
}
