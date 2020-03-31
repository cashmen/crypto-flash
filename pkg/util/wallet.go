package util

import "fmt"

type Wallet struct {
	tag string
	balances map[string] float64
}

func NewWallet() *Wallet {
	return &Wallet{
		tag: "Wallet",
		balances: make(map[string] float64),
	}
}

func (w *Wallet) Increase(coin string, amount float64) {
	if _, exist := w.balances[coin]; !exist {
		w.balances[coin] = amount
		return
	}
	w.balances[coin] += amount
}
func (w *Wallet) Decrease(coin string, amount float64) {
	if _, exist := w.balances[coin]; !exist || w.balances[coin] < amount {
		Error(w.tag, "Not enough balance for " + coin)
		return
	}
	w.balances[coin] -= amount
}
func (w *Wallet) GetBalance(coin string) float64 {
	if _, exist := w.balances[coin]; !exist {
		return 0
	}
	return w.balances[coin]
}
func (w *Wallet) String() string {
	result := ""
	for coin, balance := range w.balances {
		result += fmt.Sprintf("%s: %f ", coin, balance)
	}
	return result
}