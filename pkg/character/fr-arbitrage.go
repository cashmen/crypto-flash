/* 
// Funding Rate Arbitrage is a signal provider utilizes funding rate on 
// perpetual contract to earn profit. 
// TODO: all
*/
package character

import (
	"fmt"
	"time"
	"math"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
)

type FRArb struct {
	SignalProvider
	ftx *exchange.FTX
	// strategy config
	updatePeriod int64
	futures []string
	leverage float64
	longTime int
	aprThreshold float64
	// data
	fundingRates map[string][]float64
}
func NewFRArb(ftx *exchange.FTX, notifier *Notifier) *FRArb {
	return &FRArb{
		SignalProvider: SignalProvider{
			tag: "FRArbProvider",
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
		updatePeriod: 15,
		futures: []string{
			"BTC-PERP",
			"ETH-PERP",
		},
		leverage: 5,
		// 5 consecutive hours of positive/negative funding rate
		longTime: 5,
		aprThreshold: 0.3,
		// data
		fundingRates: make(map[string][]float64),
	}
}

func (fra *FRArb) Backtest(startTime, endTime int64) float64 {
	/*
	candles := 
		sh.ftx.GetHistoryCandles(sh.market, 300, startTime, endTime)
	util.Info(sh.tag, "start backtesting")
	for _, candle := range candles {
		sh.genSignal(candle.GetAvg(), candle.GetAvg())
	}
	roi := util.CalcROI(sh.initBalance, sh.balance)
	util.Info(sh.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", sh.balance, roi * 100))
	return roi*/
	return 0
}

func (fra *FRArb) genSignal(future string, nextFundingRate float64) {
	util.Info(fra.tag, future, 
		fmt.Sprintf("nextFundingRate: %f", nextFundingRate))
	estApr := math.Abs(nextFundingRate) * 365 * 24 * fra.leverage / 2
	util.Info(fra.tag, future, 
		fmt.Sprintf("estApr: %.2f%%", estApr * 100))
	// somehow judge if the funding rate is positive / negative for a long time
	rateHistory := fra.fundingRates[future]
	fmt.Println(rateHistory)
	// if yes => check if basis is large enough
	// if yes => calculate potential ROI
	// if good enough => open position in perp and future with leverage
	
	roi := util.CalcROI(fra.initBalance, fra.balance)
	util.Info(fra.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", fra.balance, roi * 100))
}
func (fra *FRArb) Start() {
	// get previous funding rate
	end := time.Now().Unix()
	start := end - 24 * 60 * 60
	for _, future := range fra.futures {
		fra.fundingRates[future] = fra.ftx.GetFundingRates(start, end, future)
	}
	for {
		for _, future := range fra.futures {
			// get latest funding rate
			stats := fra.ftx.GetFutureStats(future)
			fra.genSignal(future, stats.NextFundingRate)
			// if one hour just passed, append predictedRate to fundingRates
			if time.Now().Unix() % (60 * 60) == 0 {
				fra.fundingRates[future] = 
					append([]float64{stats.NextFundingRate}, 
					fra.fundingRates[future][:23]...)
			}
		}
		timeToNextCycle := 
			fra.updatePeriod - time.Now().Unix() % fra.updatePeriod
		sleepDuration := util.Duration{Second: timeToNextCycle}
		time.Sleep(sleepDuration.GetTimeDuration())
	}
}


