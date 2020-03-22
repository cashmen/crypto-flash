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
func NewTrader(ftx *exchange.FTX, notifier *Notifier) *Trader {
	w := ftx.GetWallet()
	return &Trader{
		tag: "Trader",
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
	roi := (t.wallet.GetBalance("USD") - t.initBalance) / t.initBalance
	msg := "Report\n"
	runTime := time.Now().Sub(t.startTime)
	d := util.FromTimeDuration(runTime)
	msg += "Runtime: " + d.String() + "\n"
	msg += fmt.Sprintf("Init Balance: %.2f\n", t.initBalance)
	msg += fmt.Sprintf("Balance: %.2f\n", t.wallet.GetBalance("USD"))
	msg += fmt.Sprintf("ROI: %.2f%%\n", roi * 100)
	ar := roi * (86400 * 365) / runTime.Seconds()
	msg += fmt.Sprintf("Annualized Return: %.2f%%", ar * 100)
	t.notifier.Broadcast(t.tag, msg)
}
func (t *Trader) notifyClosePosition(price, roi float64, reason string) {
	if t.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("close %s @ %.2f due to %s\n", 
		t.position.Side, price, reason)
	msg += fmt.Sprintf("ROI: %.2f%%", roi * 100)
	t.notifier.Broadcast(t.tag, msg)
	t.notifyROI()
}
func (t *Trader) notifyOpenPosition(reason string) {
	if t.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("start %s @ %.2f due to %s", 
		t.position.Side, t.position.OpenPrice, reason)
	t.notifier.Broadcast(t.tag, msg)
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
		if signal.Side == "close" {
			if t.position == nil {
				continue
			} else if t.position.Side == "short" {
				// close short position
				curMP := orderbook.Ask[0].Price
				order := &util.Order{
					Market: signal.Market,
					Side: "buy",
					Type: "market",
					Size: t.position.Size,
				}
				t.ftx.MakeOrder(order)
				util.Info(t.tag, 
					fmt.Sprintf("close short position @ %.2f", curMP))
				roi := t.position.Close(curMP)
				t.wallet = t.ftx.GetWallet()
				t.notifyClosePosition(curMP, roi, "Signal Provider")
				t.position = nil
			} else if t.position.Side == "long" {
				// close long position
				curMP := orderbook.Bid[0].Price
				util.Info(t.tag, "close long position")
				order := &util.Order{
					Market: signal.Market,
					Side: "sell",
					Type: "market",
					Size: t.position.Size,
				}
				t.ftx.MakeOrder(order)
				util.Info(t.tag, 
					fmt.Sprintf("close long position @ %.2f", curMP))
				roi := t.position.Close(curMP)
				t.wallet = t.ftx.GetWallet()
				t.notifyClosePosition(curMP, roi, signal.Reason)
				t.position = nil
			}
		} else if signal.Side == "long" {
			curMP := orderbook.Ask[0].Price
			t.wallet = t.ftx.GetWallet()
			usdBalance := t.wallet.GetBalance("USD")
			util.Info(t.tag, fmt.Sprintf("current balance: %.2f", usdBalance))
			size := usdBalance / curMP * t.leverage
			order := &util.Order{
				Market: signal.Market,
				Side: "buy",
				Type: "market",
				Size: size,
			}
			t.ftx.MakeOrder(order)
			t.position = util.NewPosition("long", size, curMP)
			util.Success(t.tag, 
				fmt.Sprintf("open postition %s with size %.4f @ %.2f",
				t.position.Side, t.position.Size, t.position.OpenPrice))
			t.notifyOpenPosition(signal.Reason)
		} else if signal.Side == "short" {
			curMP := orderbook.Bid[0].Price
			t.wallet = t.ftx.GetWallet()
			usdBalance := t.wallet.GetBalance("USD")
			util.Info(t.tag, fmt.Sprintf("current balance: %.2f", usdBalance))
			size := usdBalance / curMP * t.leverage
			order := &util.Order{
				Market: signal.Market,
				Side: "sell",
				Type: "market",
				Size: size,
			}
			t.ftx.MakeOrder(order)
			t.position = util.NewPosition("short", size, curMP)
			util.Success(t.tag, 
				fmt.Sprintf("open postition %s with size %.4f @ %.2f",
				t.position.Side, t.position.Size, t.position.OpenPrice))
			t.notifyOpenPosition(signal.Reason)
		}
	}
}
