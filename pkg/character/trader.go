/*
// Trader have the following functions:
// 1. receives trade signals
// 2. decides the amount of order to make by wallet balance
// 3. complete trades by exchange api and maintains order status
// 4. calculate ROI
// 5. send notifications such as ROI or order status
// Input: signal provider
// Output: exchange trade API or notifier
// TODO: all
*/
package character

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
