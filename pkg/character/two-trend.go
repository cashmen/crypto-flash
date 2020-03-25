/* 
// Two trend is a signal provider using two supertrends. One is used to enter.
// The other is used to take profit or stop loss.
*/
package character
/*
import (
	"fmt"
	"time"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	indicator "github.com/CheshireCatNick/crypto-flash/pkg/indicator"
)

type TwoTrend struct {
	SignalProvider
	tag string
	ftx *exchange.FTX
	startTime time.Time
	position *util.Position
	prevSide string
	initBalance float64
	balance float64
	notifier *Notifier
	signalChan chan<- *util.Signal
	mMul float64
	sMul float64
	period int
}
// strategy configuration
const (
	warmUpCandleNum = 40
	takeProfit = 200
	stopLoss = 100
	initBalance = 1000000
	market = "BTC-PERP"
	resolution = 300
)

func NewSignalProvider(ftx *exchange.FTX, notifier *Notifier) *SignalProvider {
	return &SignalProvider{
		tag: "SignalProvider",
		ftx: ftx,
		position: nil,
		prevSide: "unknown",
		initBalance: initBalance,
		balance: initBalance,
		notifier: notifier,
		signalChan: nil,
		mMul: 2.5,
		sMul: 2,
		period: 11,
	}
}
func (sp *SignalProvider) Backtest(startTime, endTime int64,
		mMul, sMul float64, period int) float64 {
	candles := 
		sp.ftx.GetHistoryCandles(market, resolution, startTime, endTime)
	st := indicator.NewSuperTrend(mMul, period)
	stopST := indicator.NewSuperTrend(sMul, period)
	
	if len(candles) <= warmUpCandleNum {
		util.Error(sp.tag, "Not enough candles for backtesting!")
	}
	for i := 0; i < warmUpCandleNum; i++ {
		st.Update(candles[i])
		stopST.Update(candles[i])
	}
	util.Info(sp.tag, "start backtesting")
	for i := warmUpCandleNum; i < len(candles); i++ {
		candle := candles[i]
		superTrend := st.Update(candle)
		stop := stopST.Update(candle)
		util.Info(sp.tag, candle.String())
		util.Info(sp.tag, util.PF64(superTrend))
		sp.genSignal(candle, superTrend, stop)
	}
	roi := (sp.balance - sp.initBalance) / sp.initBalance
	util.Info(sp.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", sp.balance, roi * 100))
	return roi
}
func (sp *SignalProvider) notifyROI() {
	if sp.notifier == nil {
		return;
	}
	roi := (sp.balance - sp.initBalance) / sp.initBalance
	msg := "Report\n"
	runTime := time.Now().Sub(sp.startTime)
	d := util.FromTimeDuration(runTime)
	msg += "Runtime: " + d.String() + "\n"
	msg += fmt.Sprintf("Init Balance: %.2f\n", sp.initBalance)
	msg += fmt.Sprintf("Balance: %.2f\n", sp.balance)
	msg += fmt.Sprintf("ROI: %.2f%%\n", roi * 100)
	ar := roi * (86400 * 365) / runTime.Seconds()
	msg += fmt.Sprintf("Annualized Return: %.2f%%", ar * 100)
	sp.notifier.Broadcast(sp.tag, msg)
}
func (sp *SignalProvider) notifyClosePosition(price, roi float64, reason string) {
	if sp.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("close %s @ %.2f due to %s\n", 
		sp.position.Side, price, reason)
	msg += fmt.Sprintf("ROI: %.2f%%", roi * 100)
	sp.notifier.Broadcast(sp.tag, msg)
	sp.notifyROI()
}
func (sp *SignalProvider) notifyOpenPosition(reason string) {
	if sp.notifier == nil {
		return;
	}
	msg := fmt.Sprintf("start %s @ %.2f due to %s", 
		sp.position.Side, sp.position.OpenPrice, reason)
	sp.notifier.Broadcast(sp.tag, msg)
}
func (sp *SignalProvider) genSignal(
		candle *util.Candle, superTrend float64, stop float64) {
	if (superTrend == -1) {
		return
	}
	
	// const take profit or stop loss
	if sp.position != nil && sp.position.Side == "long" {
		if candle.High - sp.position.OpenPrice >= takeProfit {
			price := sp.position.OpenPrice + takeProfit
			roi := sp.position.Close(price)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(price, roi, "take profit")
			sp.prevSide = sp.position.Side
			sp.position = nil
		} else if (sp.position.OpenPrice - candle.Low >= stopLoss) {
			price := sp.position.OpenPrice - stopLoss
			roi := sp.position.Close(price)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(price, roi, "stop loss")
			sp.prevSide = sp.position.Side
			sp.position = nil
		}
	} else if sp.position != nil && sp.position.Side == "short" {
		if candle.High - sp.position.OpenPrice >= stopLoss {
			price := sp.position.OpenPrice + stopLoss
			roi := sp.position.Close(price)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(price, roi, "stop loss")
			sp.prevSide = sp.position.Side
			sp.position = nil
		} else if (sp.position.OpenPrice - candle.Low >= takeProfit) {
			price := sp.position.OpenPrice - takeProfit
			roi := sp.position.Close(price)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(price, roi, "take profit")
			sp.prevSide = sp.position.Side
			sp.position = nil
		}
	}
	// dynamic take profit and stop loss by another super trend
	if sp.position != nil && sp.position.Side == "long" {
		if candle.Close <= stop {
			price := candle.Close
			roi := sp.position.Close(price)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(price, roi, "take profit or stop loss")
			sp.prevSide = sp.position.Side
			sp.position = nil
			if sp.signalChan != nil {
				sp.signalChan <- &util.Signal{ 
					Market: market, 
					Side: "close",
					Reason: "take profit or stop loss",
				}
			}
		}
	} else if sp.position != nil && sp.position.Side == "short" {
		if candle.Close >= stop {
			price := candle.Close
			roi := sp.position.Close(price)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(price, roi, "take profit or stop loss")
			sp.prevSide = sp.position.Side
			sp.position = nil
			if sp.signalChan != nil {
				sp.signalChan <- &util.Signal{ 
					Market: market, 
					Side: "close",
					Reason: "take profit or stop loss",
				}
			}
		}
	}
	if (sp.position == nil || sp.position.Side == "long") && 
			candle.Close < superTrend &&
			sp.prevSide != "short" {
		if sp.position != nil && sp.position.Side == "long" {
			// close long position
			// close price should be market price
			roi := sp.position.Close(candle.Close)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(candle.Close, roi, "SuperTrend")
			if sp.signalChan != nil {
				sp.signalChan <- &util.Signal{ 
					Market: market, 
					Side: "close",
					Reason: "SuperTrend",
				}
			}
		}
		if sp.signalChan != nil {
			sp.signalChan <- &util.Signal{ 
				Market: market, 
				Side: "short",
				Reason: "SuperTrend",
			}
		}
		sp.position = util.NewPosition("short", sp.balance, candle.Close)
		util.Info(sp.tag, 
			util.Red(fmt.Sprintf("start short @ %.2f", sp.position.OpenPrice)))
		sp.notifyOpenPosition("SuperTrend")
	} else if (sp.position == nil || sp.position.Side == "short") && 
			candle.Close > superTrend &&
			sp.prevSide != "long" {
		if sp.position != nil && sp.position.Side == "short" {
			// close short position
			// close price should be market price
			roi := sp.position.Close(candle.Close)
			sp.balance *= 1 + roi
			sp.notifyClosePosition(candle.Close, roi, "SuperTrend")
			if sp.signalChan != nil {
				sp.signalChan <- &util.Signal{ 
					Market: market, 
					Side: "close",
					Reason: "SuperTrend",
				}
			}
		}
		if sp.signalChan != nil {
			sp.signalChan <- &util.Signal{ 
				Market: market, 
				Side: "long", 
				Reason: "SuperTrend",
			}
		}
		sp.position = util.NewPosition("long", sp.balance, candle.Close)
		util.Info(sp.tag, 
			util.Green(fmt.Sprintf("start long @ %.2f", sp.position.OpenPrice)))
		sp.notifyOpenPosition("SuperTrend")
	}
	roi := (sp.balance - sp.initBalance) / sp.initBalance
	util.Info(sp.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", sp.balance, roi * 100))
}
func (sp *SignalProvider) AdjustParams() {
	ftx := exchange.NewFTX("", "", "")
	endTime := time.Now()
	d := util.Duration{ Day: -3 }
	startTime := endTime.Add(d.GetTimeDuration())
	const (
		mulMin = 1.0
		mulMax = 4.0
		pMin = 8
		pMax = 16
	)
	var bROI float64 = -100
	var bMMul, bSMul float64
	var bPeriod int
	for mMul := mulMin; mMul <= mulMax; mMul += 0.5 {
		for sMul := mulMin; sMul <= mMul; sMul += 0.5 {
			for period := pMin; period <= pMax; period++ {
				sp := NewSignalProvider(ftx, nil)
				roi := sp.Backtest(startTime.Unix(), endTime.Unix(), 
					mMul, sMul, period)
				fmt.Printf("mMul: %.1f, sMul: %.1f, period: %d, roi: %.5f\n",
					mMul, sMul, period, roi)
				if roi >= bROI {
					bROI = roi
					bMMul = mMul
					bSMul = sMul
					bPeriod = period
				}
			}
		}
	}
	fmt.Printf("mMul: %.1f, sMul: %.1f, period: %d, roi: %.2f\n",
		bMMul, bSMul, bPeriod, bROI)
	sp.mMul = bMMul
	sp.sMul = bSMul
	sp.period = bPeriod
}
func (sp *SignalProvider) Start(signalChan chan<- *util.Signal) {
	sp.signalChan = signalChan
	sp.startTime = time.Now()
	st := indicator.NewSuperTrend(sp.mMul, sp.period)
	stopST := indicator.NewSuperTrend(sp.sMul, sp.period)
	// warm up for moving average
	now := time.Now().Unix()
	resolution64 := int64(resolution)
	last := now - now % resolution64
	startTime := last - resolution64 * (warmUpCandleNum + 1) + 1
	endTime := last - resolution64
	candles := 
		sp.ftx.GetHistoryCandles(market, resolution, startTime, endTime)
	for _, candle := range candles {
		st.Update(candle)
		stopST.Update(candle)
	}
	// start real time
	c := make(chan *util.Candle)
	go sp.ftx.SubCandle(market, resolution, c);
	for {
		candle := <-c
		superTrend := st.Update(candle)
		stop := stopST.Update(candle)
		util.Info(sp.tag, "received candle", candle.String())
		util.Info(sp.tag, "super trend", util.PF64(superTrend))
		sp.genSignal(candle, superTrend, stop)
	}
}
*/