package main

import "fmt"

func (cli *CLI)reindexUTXO() {
	blockchain := NewBlockChain()
	utxoset := UTXOSet{blockchain}
	utxoset.Reindex()
	count := utxoset.CountTransactions()
	fmt.Printf("已经有%s次交易在UTXO集合\n",count)
}
