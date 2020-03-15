/* 
// Signal provider provide trade signals. A trade signal should include 
// informations such as buy/sell, stop-loss/take-profit prices, confidence, etc.
// Signal provider is an implementation of a strategy.
// Input: market data from exchanges or indicators
// Output: trader or notifier
// TODO: all
*/
package character

import (
	"fmt"
	"time"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	indicator "github.com/CheshireCatNick/crypto-flash/pkg/indicator"
)

type SignalProvider struct {
	resolution int
	market string
	position *util.Position
	prevSide string
	initBalance float64
	balance float64
}
// strategy configuration
const (
	warmUpCandleNum = 14
	takeProfit = 300
	stopLoss = 100
	initBalance = 1000000
)

func NewSignalProvider(market string, resolution int) *SignalProvider {
	return &SignalProvider{
		resolution: resolution,
		market: market,
		position: nil,
		prevSide: "unknown",
		initBalance: initBalance,
		balance: initBalance,
	}
}

func (sp *SignalProvider) Backtest(startTime, endTime int64) {
	st := indicator.NewSuperTrend(3, 10)
	ftx := exchange.NewFTX()
	candles := 
		ftx.GetHistoryCandles(sp.market, sp.resolution, startTime, endTime)
	for i := 0; i < warmUpCandleNum; i++ {
		st.Update(candles[i])
	}
	fmt.Println("start backtesting")
	for i := warmUpCandleNum; i < len(candles); i++ {
		candle := candles[i]
		superTrend := st.Update(candle)
		fmt.Println("candle:", candle.High, candle.Low, candle.Close, candle.StartTime)
		fmt.Println("supertrend:", st.Update(candle))
		sp.genSignal(candle, superTrend)
	}
	ROI := (sp.balance - sp.initBalance) / sp.initBalance
	fmt.Printf("balance: %f, total ROI: %f\n", sp.balance, ROI)
}
func (sp *SignalProvider) genSignal(candle *util.Candle, superTrend float64) {
	fmt.Printf("close %f, st: %f\n", candle.Close, superTrend)
	// take profit or stop loss
	if sp.position != nil && sp.position.Side == "long" {
		if candle.High - sp.position.OpenPrice >= takeProfit {
			ROI := sp.position.Close(sp.position.OpenPrice + takeProfit)
			sp.balance *= 1 + ROI
			sp.prevSide = sp.position.Side
			sp.position = nil
		} else if (sp.position.OpenPrice - candle.Low >= stopLoss) {
			ROI := sp.position.Close(sp.position.OpenPrice - stopLoss)
			sp.balance *= 1 + ROI
			sp.prevSide = sp.position.Side
			sp.position = nil
		}
	} else if sp.position != nil && sp.position.Side == "short" {
		if candle.High - sp.position.OpenPrice >= stopLoss {
			ROI := sp.position.Close(sp.position.OpenPrice + stopLoss)
			sp.balance *= 1 + ROI
			sp.prevSide = sp.position.Side
			sp.position = nil
		} else if (sp.position.OpenPrice - candle.Low >= takeProfit) {
			ROI := sp.position.Close(sp.position.OpenPrice - takeProfit)
			sp.balance *= 1 + ROI
			sp.prevSide = sp.position.Side
			sp.position = nil
		}
	}
	if (sp.position == nil || sp.position.Side == "long") && 
			candle.Close < superTrend &&
			sp.prevSide != "short" {
		if sp.position != nil && sp.position.Side == "long" {
			// close long position
			// close price should be market price
			ROI := sp.position.Close(candle.Close)
			sp.balance *= 1 + ROI
		}
		sp.position = util.NewPosition("short", sp.balance, candle.Close)
		fmt.Printf("start short @ %f\n", sp.position.OpenPrice)
	} else if (sp.position == nil || sp.position.Side == "short") && 
			candle.Close > superTrend &&
			sp.prevSide != "long" {
		if sp.position != nil && sp.position.Side == "short" {
			// close short position
			// close price should be market price
			ROI := sp.position.Close(candle.Close)
			sp.balance *= 1 + ROI
		}
		sp.position = util.NewPosition("long", sp.balance, candle.Close)
		fmt.Printf("start long @ %f\n", sp.position.OpenPrice)
	}
	ROI := (sp.balance - sp.initBalance) / sp.initBalance
	fmt.Printf("balance: %f, total ROI: %f\n", sp.balance, ROI)
}
func (sp *SignalProvider) Start() {
	st := indicator.NewSuperTrend(3, 10)
	ftx := exchange.NewFTX()
	// warm up for moving average
	now := time.Now().Unix()
	resolution64 := int64(sp.resolution)
	last := now - now % resolution64
	startTime := last - resolution64 * (warmUpCandleNum + 1) + 1
	endTime := last - resolution64
	candles := 
		ftx.GetHistoryCandles(sp.market, sp.resolution, startTime, endTime)
	for _, candle := range candles {
		st.Update(candle)
	}
	// start real time
	c := make(chan *util.Candle)
	go ftx.SubCandle(sp.market, sp.resolution, c);
	for {
		candle := <-c
		superTrend := st.Update(candle)
		fmt.Println("candle:", candle.High, candle.Low, candle.Close, candle.StartTime)
		fmt.Println("supertrend:", st.Update(candle))
		sp.genSignal(candle, superTrend)
	}
}