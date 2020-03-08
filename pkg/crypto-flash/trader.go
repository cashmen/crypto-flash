package cryptoflash

import "fmt"

// Trader represent a trader in crypto flash
type Trader struct {
	id string
}

// NewTrader creates a trader instance
func NewTrader() *Trader {

	return &Trader{id: "1234"}
}

// Run starts a trader
func (t *Trader) Run() {
	fmt.Println("start trading")
}
