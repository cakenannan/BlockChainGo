package main

func main() {
	chain := NewBlockChain() //创建区块链
	defer chain.db.Close()   //延迟关闭数据库
	cli := CLI{chain}		//创建命令行
	cli.Run()

	//在gamestart文件夹下打开cmd,执行	go build ./	,得到gamestart.exe
	//gamestart.exe showchain
	//gamestart.exe addblock -data "data2"
}
