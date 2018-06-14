package main

import (
	"log"
	"fmt"
)

func (cli *CLI) listAddress() {
	wallets,err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for addr := range addresses {
		fmt.Println(addr)
	}
}
