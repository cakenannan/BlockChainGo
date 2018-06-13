package main

func main() {
	cli := CLI{}		//创建命令行
	cli.Run()

	//在gamestart文件夹下打开cmd,执行	go build ./	,得到gamestart.exe
	//gamestart.exe showchain
	//gamestart.exe addblock -data "data2"


}
