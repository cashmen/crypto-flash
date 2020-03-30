/*
// Trader have the following functions:
// 1. receives trade signals
// 2. decides the amount of order to make by wallet balance
// 3. complete trades by exchange api and maintains order status
// 4. calculate ROI
// 5. send notifications such as ROI or order status
// Input: signal provider
// Output: exchange trade API or notifier
// TODO:
// 1. track order
// 2. precautions
*/
package character

import (
	"time"
	"fmt"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
)

// Trader represent a trader in crypto flash
type Trader struct {
	tag string
	name string
	startTime time.Time
	ftx *exchange.FTX
	notifier *Notifier
	wallet *util.Wallet
	position *util.Position
	ignoreFirstSignal bool
	initBalance float64
	leverage float64
}

// NewTrader creates a trader instance
func NewTrader(name string, ftx *exchange.FTX, notifier *Notifier) *Trader {
	w := ftx.GetWallet()
	return &Trader{
		tag: "Trader-" + name,
		name: name,
		ftx: ftx,
		notifier: notifier,
		wallet: w,
		initBalance: w.GetBalance("USD"),
		position: nil,
		// ignore first signal?
		ignoreFirstSignal: false,
		leverage: 1,
	}
}
func (t *Trader) notifyROI() {
	if t.notifier == nil {
		return;
	}
	roi := util.CalcROI(t.initBalance, t.wallet.GetBalance("USD"))
	msg := "Report\n"
	runTime := time.Now().Sub(t.startTime)
	d := util.FromTimeDuration(runTime)
	msg += "Runtime: " + d.String() + "\n"
	msg += fmt.Sprintf("Init Balance: %.2f\n", t.initBalance)
	msg += fmt.Sprintf("Balance: %.2f\n", t.wallet.GetBalance("USD"))
	msg += fmt.Sprintf("ROI: %.2f%%\n", roi * 100)
	ar := roi * (86400 * 365) / runTime.Seconds()
	msg += fmt.Sprintf("Annualized Return: %.2f%%", ar * 100)
	t.notifier.Send(t.tag, t.name, msg)
}
func (t *Trader) notifyClosePosition(price, roi float64, reason string) {
	if t.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("close %s @ %.2f due to %s\n", 
		t.position.Side, price, reason)
	msg += fmt.Sprintf("ROI: %.2f%%", roi * 100)
	t.notifier.Send(t.tag, t.name, msg)
	t.notifyROI()
}
func (t *Trader) notifyOpenPosition(reason string) {
	if t.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("start %s @ %.2f due to %s", 
		t.position.Side, t.position.OpenPrice, reason)
	t.notifier.Send(t.tag, t.name, msg)
}
func (t *Trader) closePosition(market string, price float64, reason string) {
	action := ""
	if t.position.Side == "short" {
		action = "buy"
	} else if t.position.Side == "long" {
		action = "sell"
	}
	order := &util.Order{
		Market: market,
		Side: action,
		Type: "market",
		Size: t.position.Size,
	}
	t.ftx.MakeOrder(order)
	roi := t.position.Close(price)
	t.wallet = t.ftx.GetWallet()
	t.notifyClosePosition(price, roi, reason)
	logMsg := fmt.Sprintf("close %s @ %.2f due to %s, ROI: %.2f%%", 
		t.position.Side, price, reason, roi * 100)
	if roi > 0 { 
		util.Info(t.tag, util.Green(logMsg))
	} else {
		util.Info(t.tag, util.Red(logMsg))
	}
	t.position = nil
}
func (t *Trader) openPosition(
		market, side string, size, price float64, reason string) {
	action := ""
	if side == "long" {
		action = "buy"
	} else if side == "short" {
		action = "sell"
	}
	order := &util.Order{
		Market: market,
		Side: action,
		Type: "market",
		Size: size,
	}
	t.ftx.MakeOrder(order)
	t.position = util.NewPosition(side, size, price)
	t.notifyOpenPosition(reason)
	logMsg := fmt.Sprintf("start %s @ %.2f due to %s", side, price, reason)
	if side == "long" {
		util.Info(t.tag, util.Green(logMsg))
	} else {
		util.Info(t.tag, util.Red(logMsg))
	}
}
// Run starts a trader
func (t *Trader) Start(signalChan <-chan *util.Signal) {
	t.startTime = time.Now()
	for signal := range signalChan {
		util.Info(t.tag, "receive signal: " + signal.Side)
		// ignore the first signal
		if t.ignoreFirstSignal {
			t.ignoreFirstSignal = false
			continue
		}
		orderbook := t.ftx.GetOrderbook(signal.Market, 1)
		var curMP float64
		if signal.Side == "close" {
			if t.position == nil {
				continue
			}
			if t.position.Side == "short" {
				curMP = orderbook.Ask[0].Price
			} else if t.position.Side == "long" {
				curMP = orderbook.Bid[0].Price
			}
			go t.closePosition(signal.Market, curMP, signal.Reason)	
		} else if signal.Side == "long" || signal.Side == "short" {
			if signal.Side == "long" {
				curMP = orderbook.Ask[0].Price
			} else if signal.Side == "short" {
				curMP = orderbook.Bid[0].Price
			}
			usdBalance := t.wallet.GetBalance("USD")
			util.Info(t.tag, fmt.Sprintf("current balance: %.2f", usdBalance))
			size := usdBalance / curMP * t.leverage
			go t.openPosition(
				signal.Market, signal.Side, size, curMP, signal.Reason)
		}
	}
}
