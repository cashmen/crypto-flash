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
// 3. track free balance (fund pool)
*/
package character

import (
	"time"
	"fmt"
	exchange "github.com/cashmen/crypto-flash/pkg/exchange"
	util "github.com/cashmen/crypto-flash/pkg/util"
)

// Trader represent a trader in crypto flash
type Trader struct {
	tag string
	owner string
	startTime time.Time
	ftx *exchange.FTX
	notifier *Notifier
	wallet *util.Wallet
	market string
	position *util.Position
	curPosition *util.Position
	ignoreFirstSignal bool
	initBalance float64
	leverage float64
	updatePeriod time.Duration
	stopLoss float64
	takeProfit float64
}

// NewTrader creates a trader instance
func NewTrader(owner string, ftx *exchange.FTX, notifier *Notifier) *Trader {
	w := ftx.GetWallet()
	util.Success("Trader-" + owner, "successfully get balance", w.String())
	t := &Trader{
		tag: "Trader-" + owner,
		owner: owner,
		ftx: ftx,
		notifier: notifier,
		wallet: w,
		initBalance: w.GetBalance("ETH"),
		market: "SHIT-PERP",
		position: nil,
		// ignore first signal?
		ignoreFirstSignal: false,
		leverage: 1,
		updatePeriod: 10 * 60 * time.Second,
	}
	go t.updateStatus()
	return t;
}
func (t *Trader) notifyROI() {
	if t.notifier == nil {
		return;
	}
	t.wallet = t.ftx.GetWallet()
	roi := util.CalcROI(t.initBalance, t.wallet.GetBalance("ETH"))
	msg := "Report\n"
	runTime := time.Now().Sub(t.startTime)
	d := util.FromTimeDuration(runTime)
	msg += "Runtime: " + d.String() + "\n"
	msg += fmt.Sprintf("Init Balance: %.2f\n", t.initBalance)
	msg += fmt.Sprintf("Balance: %.2f\n", t.wallet.GetBalance("ETH"))
	msg += fmt.Sprintf("ROI: %.2f%%\n", roi * 100)
	ar := roi * (86400 * 365) / runTime.Seconds()
	msg += fmt.Sprintf("Annualized Return: %.2f%%", ar * 100)
	t.notifier.Send(t.tag, t.owner, msg)
}
func (t *Trader) notifyClosePosition(price, roi float64, reason string) {
	if t.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("close %s @ %.2f due to %s\n", 
		t.position.Side, price, reason)
	msg += fmt.Sprintf("ROI: %.2f%%", roi * 100)
	t.notifier.Send(t.tag, t.owner, msg)
	t.notifyROI()
}
func (t *Trader) notifyOpenPosition(reason string) {
	if t.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("start %s @ %.2f due to %s", 
		t.position.Side, t.position.OpenPrice, reason)
	t.notifier.Send(t.tag, t.owner, msg)
}
func (t *Trader) updateStatus() {
	for {
		t.wallet = t.ftx.GetWallet()
		util.Success(t.tag, "successfully update balance", t.wallet.String())
		t.curPosition = t.ftx.GetPosition(t.market)
		if t.curPosition != nil {
			util.Success(t.tag, "successfully update position", 
				t.curPosition.String())
		} else {
			util.Success(t.tag, "no current position")
		}
		time.Sleep(t.updatePeriod)
	}
}
func (t *Trader) closePosition(market string, price float64, reason string) {
	t.curPosition = t.ftx.GetPosition(market)
	if t.curPosition != nil {
		action := ""
		if t.curPosition.Side == "short" {
			action = "buy"
		} else if t.curPosition.Side == "long" {
			action = "sell"
		}
		order := &util.Order{
			Market: market,
			Side: action,
			Type: "market",
			Size: t.curPosition.Size,
			ReduceOnly: true,
		}
		t.ftx.MakeOrder(order)
	}
	if reason == "take profit" {
		price = t.takeProfit
	} else if reason == "stop loss" {
		price = t.stopLoss
	}
	roi := t.position.Close(price)
	t.notifyClosePosition(price, roi, reason)
	logMsg := fmt.Sprintf("close %s @ %.2f due to %s, ROI: %.2f%%", 
		t.position.Side, price, reason, roi * 100)
	if roi > 0 {
		util.Info(t.tag, util.Green(logMsg))
	} else {
		util.Info(t.tag, util.Red(logMsg))
	}
	t.position = nil
	t.ftx.CancelAllOrder(market)
}
func (t *Trader) openPosition(signal *util.Signal, size, price float64) {
	var action, exitAction string
	if signal.Side == "long" {
		action = "buy"
		exitAction = "sell"
	} else if signal.Side == "short" {
		action = "sell"
		exitAction = "buy"
	}
	order := &util.Order{
		Market: signal.Market,
		Side: action,
		Type: "market",
		Size: size,
	}
	t.ftx.MakeOrder(order)
	t.position = util.NewPosition(signal.Side, size, price)
	t.notifyOpenPosition(signal.Reason)
	logMsg := fmt.Sprintf("start %s @ %.2f due to %s", 
		signal.Side, price, signal.Reason)
	if signal.Side == "long" {
		util.Info(t.tag, util.Green(logMsg))
	} else {
		util.Info(t.tag, util.Red(logMsg))
	}
	if signal.TakeProfit > 0 {
		/*
		takeProfitOrder := &util.Order{
			Market: signal.Market,
			Side: exitAction,
			Size: size,
			Type: "takeProfit",
			ReduceOnly: true,
			RetryUntilFilled: true,
			TriggerPrice: signal.TakeProfit,
			//OrderPrice: 6500,
		}*/
		takeProfitOrder := &util.Order{
			Market: signal.Market,
			Side: exitAction,
			Type: "limit",
			Size: size,
			ReduceOnly: true,
			Price: signal.TakeProfit,
		}
		t.ftx.MakeOrder(takeProfitOrder)
		t.takeProfit = signal.TakeProfit
	}
	if signal.StopLoss > 0 {
		var order *util.Order
		if signal.UseTrailingStop {
			order = &util.Order{
				Market: signal.Market,
				Side: exitAction,
				Size: size,
				Type: "trailingStop",
				ReduceOnly: true,
				RetryUntilFilled: true,
				TrailValue: signal.StopLoss - signal.Open,
			}
		} else {
			order = &util.Order{
				Market: signal.Market,
				Side: exitAction,
				Size: size,
				Type: "stop",
				ReduceOnly: true,
				RetryUntilFilled: true,
				TriggerPrice: signal.StopLoss,
				//OrderPrice: signal.StopLoss,
			}
			t.stopLoss = signal.StopLoss
		}
		t.ftx.MakeOrder(order)
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
			usdBalance := t.wallet.GetBalance("ETH")
			util.Info(t.tag, fmt.Sprintf("current balance: %.2f", usdBalance))
			size := t.initBalance / curMP * t.leverage
			go t.openPosition(signal, size, curMP)
		}
	}
}
