package main

import (
	"fmt"

	cryptoflash "github.com/CheshireCatNick/crypto-flash/pkg"
)

func main() {
	fmt.Printf("Hello, world\n")
	trader := cryptoflash.NewTrader()
	trader.Run()
}
