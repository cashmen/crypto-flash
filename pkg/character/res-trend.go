/* 
// Resolution Trend is a signal provider utilize two supertrends from different 
// resolution.
*/
package character

import (
	"fmt"
	"time"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	indicator "github.com/CheshireCatNick/crypto-flash/pkg/indicator"
)

type ResTrend struct {
	SignalProvider
	ftx *exchange.FTX
	// strategy config
	market string
	mul float64
	res int
	mainMul float64
	mainRes int
	period int
	st *indicator.Supertrend
	mainST *indicator.Supertrend
	warmUpCandleNum int
	prevSupertrend float64
	prevMainSupertrend float64
	trend string
	prevTrend string
	mainTrend string
	prevMainTrend string
	takeProfit float64
	stopLoss float64
}
func NewResTrend(ftx *exchange.FTX, notifier *Notifier) *ResTrend {
	return &ResTrend{
		SignalProvider: SignalProvider{
			tag: "ResTrendProvider",
			startTime: time.Now(),
			position: nil,
			initBalance: 1000000,
			balance: 1000000,
			notifier: notifier,
			signalChan: nil,
		},
		ftx: ftx,
		market: "BTC-PERP",
		mul: 1,
		res: 300, // 15 (for test), 60, 300 or 900
		mainMul: 1,
		mainRes: 3600, // 60 (for test), 3600 or 14400
		period: 1,
		warmUpCandleNum: 40,
		takeProfit: 200,
		stopLoss: 100,
	}
}
func (rt *ResTrend) Backtest(startTime, endTime int64) float64 {
	candles := 
		rt.ftx.GetHistoryCandles(rt.market, rt.res, startTime, endTime)
	mainCandles :=
		rt.ftx.GetHistoryCandles(rt.market, rt.mainRes, startTime, endTime)
	if len(candles) == 5000 {
		util.Error(rt.tag, "Can't have more candles.")
	}
	rt.warmUp(startTime)
	util.Info(rt.tag, "start backtesting")
	currentMainI := 0
	for _, candle := range candles {
		if currentMainI < len(mainCandles) &&
				candle.StartTime == mainCandles[currentMainI].StartTime {
			rt.genSignal(mainCandles[currentMainI], rt.mainRes)
			currentMainI++
		}
		rt.genSignal(candle, rt.res)
	}
	roi := util.CalcROI(rt.initBalance, rt.balance)
	util.Info(rt.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", rt.balance, roi * 100))
	return roi
}
func (rt *ResTrend) genSignal(candle *util.Candle, res int) {
	var supertrend, mainSupertrend float64
	if res == rt.res {
		supertrend = rt.st.Update(candle)
		util.Info(rt.tag, "received candle", candle.String())
		util.Info(rt.tag, "supertrend", util.PF64(supertrend))
		if candle.Close > supertrend {
			rt.trend = "bull"
		} else if candle.Close < supertrend {
			rt.trend = "bear"
		}
	} else if res == rt.mainRes {
		mainSupertrend = rt.mainST.Update(candle)
		util.Info(rt.tag, "received main candle", candle.String())
		util.Info(rt.tag, "main supertrend", util.PF64(mainSupertrend))
		if candle.Close > mainSupertrend {
			rt.mainTrend = "bull"
		} else if candle.Close < mainSupertrend {
			rt.mainTrend = "bear"
		}
	}
	if (rt.trend == "" || rt.prevTrend == "" || 
			rt.mainTrend == "" || rt.prevMainTrend == "") {
		return
	}
	util.Info(rt.tag, "prevTrend:", rt.prevTrend)
	util.Info(rt.tag, "trend:", rt.trend)
	util.Info(rt.tag, "prevMainTrend:", rt.prevMainTrend)
	util.Info(rt.tag, "mainTrend:", rt.mainTrend)
	// const take profit or stop loss
	if rt.position != nil && rt.position.Side == "long" {
		if candle.High - rt.position.OpenPrice >= rt.takeProfit {
			price := rt.position.OpenPrice + rt.takeProfit
			rt.closePosition(price, "take profit")
		} else if (rt.position.OpenPrice - candle.Low >= rt.stopLoss) {
			price := rt.position.OpenPrice - rt.stopLoss
			rt.closePosition(price, "stop loss")
		}
	} else if rt.position != nil && rt.position.Side == "short" {
		if candle.High - rt.position.OpenPrice >= rt.stopLoss {
			price := rt.position.OpenPrice + rt.stopLoss
			rt.closePosition(price, "stop loss")
		} else if (rt.position.OpenPrice - candle.Low >= rt.takeProfit) {
			price := rt.position.OpenPrice - rt.takeProfit
			rt.closePosition(price, "take profit")
		}
	}
	/*
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
	}*/
	if (rt.position == nil || rt.position.Side == "long") && 
			(rt.trend == "bear" && rt.prevTrend == "bull" && 
			rt.mainTrend == "bear") {
		if rt.position != nil && rt.position.Side == "long" {
			// close long position
			// close price should be market price
			rt.closePosition(candle.Close, "Supertrend")
			/*
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "Supertrend",
			})*/
		}
		/*
		rt.sendSignal(&util.Signal{ 
			Market: rt.market, 
			Side: "short",
			Reason: "Supertrend",
		})*/
		rt.openPosition("short", rt.balance, candle.Close, "Supertrend")
	} else if (rt.position == nil || rt.position.Side == "short") && 
				(rt.trend == "bull" && rt.prevTrend == "bear" && 
				rt.mainTrend == "bull") {
		if rt.position != nil && rt.position.Side == "short" {
			// close short position
			// close price should be market price
			rt.closePosition(candle.Close, "Supertrend")
			/*
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "Supertrend",
			})*/
		}
		/*
		rt.sendSignal(&util.Signal{ 
			Market: rt.market, 
			Side: "long",
			Reason: "Supertrend",
		})*/
		rt.openPosition("long", rt.balance, candle.Close, "Supertrend")
	}
	roi := util.CalcROI(rt.initBalance, rt.balance)
	util.Info(rt.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", rt.balance, roi * 100))
	if res == rt.res {
		rt.prevSupertrend = supertrend
		rt.prevTrend = rt.trend
	} else if res == rt.mainRes {
		rt.prevMainSupertrend = mainSupertrend
		rt.prevMainTrend = rt.mainTrend
	}
}/*
func (rt *ResTrend) AdjustParams() {
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
}*/
func (rt *ResTrend) getCandles(from int64, res int) []*util.Candle {
	res64 := int64(res)
	last := from - from % res64
	startTime := last - res64 * (int64(rt.warmUpCandleNum) + 1) + 1
	endTime := last - res64
	return rt.ftx.GetHistoryCandles(rt.market, res, startTime, endTime)
}
func (rt *ResTrend) warmUp(from int64) {
	rt.st = indicator.NewSupertrend(rt.mul, rt.period)
	rt.st.Tag = "Supertrend"
	rt.mainST = indicator.NewSupertrend(rt.mainMul, rt.period)
	rt.mainST.Tag = "Main Supertrend"
	candles := rt.getCandles(from, rt.res)
	for _, candle := range candles {
		rt.prevSupertrend = rt.st.Update(candle)
		rt.prevTrend = rt.trend
		if candle.Close > rt.prevSupertrend {
			rt.trend = "bull"
		} else if candle.Close < rt.prevSupertrend {
			rt.trend = "bear"
		}
	}
	candles = rt.getCandles(from, rt.mainRes)
	for _, candle := range candles {
		rt.prevMainSupertrend = rt.mainST.Update(candle)
		rt.prevMainTrend = rt.mainTrend
		if candle.Close > rt.prevMainSupertrend {
			rt.mainTrend = "bull"
		} else if candle.Close < rt.prevMainSupertrend {
			rt.mainTrend = "bear"
		}
	}
}
func (rt *ResTrend) Start(signalChan chan<- *util.Signal) {
	rt.signalChan = signalChan
	rt.warmUp(time.Now().Unix())
	candleChan := make(chan *util.Candle)
	mainCandleChan := make(chan *util.Candle)
	go rt.ftx.SubCandle(rt.market, rt.res, candleChan);
	go rt.ftx.SubCandle(rt.market, rt.mainRes, mainCandleChan);
	for {
		select {
		case candle := <-candleChan:
			rt.genSignal(candle, rt.res)
		case candle := <-mainCandleChan:
			rt.genSignal(candle, rt.mainRes)
		}
	}
}
