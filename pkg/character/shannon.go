/* 
// Shannon is a signal provider utilize constant rebalancing portfolio to gain 
// profit.
// TODO:
*/
package character

import (
	"fmt"
	"time"
	//"math"
	exchange "github.com/cashmen/crypto-flash/pkg/exchange"
	util "github.com/cashmen/crypto-flash/pkg/util"
)

type Shannon struct {
	SignalProvider
	ftx *exchange.FTX
	// strategy config
	market string
	updatePeriod time.Duration
	threshold float64
	// data
	assets float64
	usd float64
	opCount int
}
func NewShannon(ftx *exchange.FTX, notifier *Notifier) *Shannon {
	return &Shannon{
		SignalProvider: SignalProvider{
			tag: "ShannonProvider",
			startTime: time.Now(),
			position: nil,
			initBalance: 1000000,
			balance: 1000000,
			notifier: notifier,
			signalChan: nil,
			takeProfitCount: 0,
			stopLossCount: 0,
		},
		ftx: ftx,
		// config
		market: "BTC/USD",
		updatePeriod: 5 * time.Second,
		threshold: 0.05,
		// data
		assets: 0,
		usd: 1000000,
		opCount: 0,
	}
}

func (sh *Shannon) Backtest(startTime, endTime int64) float64 {
	candles := 
		sh.ftx.GetHistoryCandles(sh.market, 300, startTime, endTime)
	util.Info(sh.tag, "start backtesting")
	for _, candle := range candles {
		sh.genSignal(candle.GetAvg(), candle.GetAvg())
	}
	roi := util.CalcROI(sh.initBalance, sh.balance)
	util.Info(sh.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", sh.balance, roi * 100))
	return roi
}

func (sh *Shannon) genSignal(ask, bid float64) {
	util.Info(sh.tag, "received price:", util.PF64(ask), util.PF64(bid))
	avgPrice := (ask + bid) / 2
	assetsWorth := sh.assets * avgPrice
	diff := sh.usd - assetsWorth
	if diff / sh.usd >= sh.threshold {
		// buy assets
		buyAmount := diff / (2 * ask)
		sh.assets += buyAmount
		sh.usd -= diff / 2
		sh.opCount++
	} else if diff / sh.usd <= -sh.threshold {
		// sell assets
		sellAmount := -diff / (2 * bid)
		sh.assets -= sellAmount
		sh.usd += -diff / 2
		sh.opCount++
	}
	assetsWorth = sh.assets * avgPrice
	diff = sh.usd - assetsWorth
	util.Info(sh.tag, fmt.Sprintf("after balancing, assets worth: %f, usd: %f", 
		assetsWorth, sh.usd))
	util.Info(sh.tag, fmt.Sprintf("difference: %.2f%%", diff * 100 / sh.usd))
	sh.balance = assetsWorth + sh.usd
	roi := util.CalcROI(sh.initBalance, sh.balance)
	util.Info(sh.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", sh.balance, roi * 100))
	util.Info(sh.tag, "operation count:", util.PI(sh.opCount))
}
func (sh *Shannon) Start() {
	for {
		orderbook := sh.ftx.GetOrderbook(sh.market, 1)
		sh.genSignal(orderbook.Ask[0].Price, orderbook.Bid[0].Price)
		time.Sleep(sh.updatePeriod)
	}
}
