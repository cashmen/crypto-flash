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

import (
	"fmt"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
)

// Trader represent a trader in crypto flash
type Trader struct {
	
}

// NewTrader creates a trader instance
func NewTrader(ftx *exchange.FTX, notifier *Notifier) *Trader {
	return &Trader{}
}

// Run starts a trader
func (t *Trader) Start() {
	fmt.Println("start trading")
}
