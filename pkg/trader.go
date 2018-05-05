package cryptoflash

import "fmt"

type Trader struct {
	id string
}

func NewTrader() *Trader {

	return &Trader{id: "1234"}
}

func (t *Trader) Run() {
	fmt.Println("start trading")
}
