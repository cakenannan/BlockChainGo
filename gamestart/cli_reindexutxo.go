package main

import "fmt"

func (cli *CLI)reindexUTXO(nodeID string) {
	blockchain := NewBlockChain(nodeID)
	utxoset := UTXOSet{blockchain}
	utxoset.Reindex()
	count := utxoset.CountTransactions()
	fmt.Printf("已经有%s次交易在UTXO集合\n",count)
}
