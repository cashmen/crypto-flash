/* 
// Resolution Trend is a signal provider utilizes two supertrends from different 
// resolution.
// TODO:
*/
package character

import (
	"fmt"
	"time"
	"math"
	exchange "github.com/cashmen/crypto-flash/pkg/exchange"
	util "github.com/cashmen/crypto-flash/pkg/util"
	indicator "github.com/cashmen/crypto-flash/pkg/indicator"
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
	warmUpCandleNum int
	takeProfit float64
	stopLoss float64
	useTrailingStop bool
	// data
	st *indicator.Supertrend
	mainST *indicator.Supertrend
	prevSupertrend float64
	trend string
	prevTrend string
	mainTrend string
	prevMainTrend string
	mainCandle *util.Candle
	stopLossPrice float64
	takeProfitPrice float64
}
func NewResTrend(ftx *exchange.FTX, notifier *Notifier) *ResTrend {
	return &ResTrend{
		SignalProvider: SignalProvider{
			tag: "ResTrendProvider",
			startTime: time.Now(),
			position: nil,
			initBalance: 1000,
			balance: 1000,
			notifier: notifier,
			signalChan: nil,
			takeProfitCount: 0,
			stopLossCount: 0,
		},
		ftx: ftx,
		// config
		market: "SHIT-PERP",
		mul: 1,
		res: 900, // 15 (for test), 60, 300 or 900
		mainMul: 2,
		mainRes: 14400, // 60 (for test), 3600 or 14400
		period: 3,
		warmUpCandleNum: 50,
		takeProfit: 11,
		stopLoss: 12,
		useTrailingStop: false,
		// data
		mainCandle: nil,
	}
}
func (rt *ResTrend) Backtest(startTime, endTime int64) float64 {
	candles := 
		rt.ftx.GetHistoryCandles(rt.market, rt.res, startTime, endTime)
	rt.warmUp(startTime)
	util.Info(rt.tag, "start backtesting")
	for _, candle := range candles {
		rt.genSignal(candle)
	}
	roi := util.CalcROI(rt.initBalance, rt.balance)
	util.Info(rt.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", rt.balance, roi * 100))
	winRate := float64(rt.takeProfitCount) / 
		float64(rt.takeProfitCount + rt.stopLossCount)
	util.Info(rt.tag, fmt.Sprintf("win rate: %.2f%%", winRate * 100))
	rt.showChart()
	return roi
}
func (rt *ResTrend) genSignal(candle *util.Candle) {
	util.Info(rt.tag, "received candle", candle.String())
	var supertrend, mainSupertrend float64
	supertrend = rt.st.Update(candle)
	if candle.GetTime().Unix() % int64(rt.mainRes) == 0 {
		rt.mainCandle = candle
	} else {
		rt.mainCandle.Update(candle)
	}
	if (candle.GetTime().Unix() + int64(rt.res)) % 
			int64(rt.mainRes) == 0 {
		util.Info(rt.tag, "main candle", rt.mainCandle.String())
		mainSupertrend = rt.mainST.Update(rt.mainCandle)
		if candle.Close > mainSupertrend {
			rt.mainTrend = "bull"
		} else if candle.Close < mainSupertrend {
			rt.mainTrend = "bear"
		}
	} else {
		mainSupertrend = rt.mainST.Predict(rt.mainCandle)
		if candle.Close > mainSupertrend {
			rt.mainTrend = "bull"
		} else if candle.Close < mainSupertrend {
			rt.mainTrend = "bear"
		}
	}
	if candle.Close > supertrend {
		rt.trend = "bull"
	} else if candle.Close < supertrend {
		rt.trend = "bear"
	}
	if (rt.trend == "" || rt.prevTrend == "" || rt.mainTrend == "") {
		return
	}
	util.Info(rt.tag, 
		fmt.Sprintf("st: %f, mainST: %f", supertrend, mainSupertrend))
	color := func(trend string) string {
		if trend == "bull" {
			return util.Green(trend)
		} else {
			return util.Red(trend)
		}
	}
	util.Info(rt.tag, "prevTrend:", color(rt.prevTrend))
	util.Info(rt.tag, "trend:", color(rt.trend))
	util.Info(rt.tag, "mainTrend:", color(rt.mainTrend))
	util.Info(rt.tag, "prevMainTrend:", color(rt.prevMainTrend))
	// const take profit or stop loss
	if rt.position != nil && rt.position.Side == "long" {
		if rt.useTrailingStop {
			rt.stopLossPrice = 
				math.Max(rt.stopLossPrice, candle.High - rt.stopLoss)
			util.Info(rt.tag, "current stop loss:", util.PF64(rt.stopLossPrice))
		}
		if candle.High >= rt.takeProfitPrice {
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "take profit",
			})
			rt.closePosition(rt.takeProfitPrice, "take profit")
		} else if (candle.Low <= rt.stopLossPrice) {
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "stop loss",
			})
			rt.closePosition(rt.stopLossPrice, "stop loss")
		}
	} else if rt.position != nil && rt.position.Side == "short" {
		if rt.useTrailingStop {
			rt.stopLossPrice = 
				math.Min(rt.stopLossPrice, candle.Low + rt.stopLoss)
			util.Info(rt.tag, "current stop loss:", util.PF64(rt.stopLossPrice))
		}
		if candle.High >= rt.stopLossPrice {
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "stop loss",
			})
			rt.closePosition(rt.stopLossPrice, "stop loss")
		} else if (candle.Low <= rt.takeProfitPrice) {
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "take profit",
			})
			rt.closePosition(rt.takeProfitPrice, "take profit")
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
			((rt.prevTrend == "bull" && rt.trend == "bear" && rt.mainTrend == "bear") ||
			(rt.prevMainTrend == "bull" && rt.mainTrend == "bear")) {
		if rt.position != nil && rt.position.Side == "long" {
			// close long position
			// close price should be market price
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "Supertrend",
			})
			rt.closePosition(candle.Close, "Supertrend")
		}
		rt.takeProfitPrice = candle.Close - rt.takeProfit
		rt.stopLossPrice = candle.Close + rt.stopLoss
		rt.sendSignal(&util.Signal{ 
			Market: rt.market, 
			Side: "short",
			Reason: "Supertrend",
			Open: candle.Close * 0.99,
			TakeProfit: rt.takeProfitPrice,
			StopLoss: rt.stopLossPrice,
			UseTrailingStop: rt.useTrailingStop,
			Ratio: 1,
		})
		rt.openPosition("short", rt.initBalance, candle.Close, "Supertrend")
	} else if (rt.position == nil || rt.position.Side == "short") && 
			((rt.prevTrend == "bear" && rt.trend == "bull" && rt.mainTrend == "bull") ||
			(rt.prevMainTrend == "bear" && rt.mainTrend == "bull")) {
		if rt.position != nil && rt.position.Side == "short" {
			// close short position
			// close price should be market price
			rt.sendSignal(&util.Signal{ 
				Market: rt.market, 
				Side: "close",
				Reason: "Supertrend",
			})
			rt.closePosition(candle.Close, "Supertrend")
		}
		rt.takeProfitPrice = candle.Close + rt.takeProfit
		rt.stopLossPrice = candle.Close - rt.stopLoss
		rt.sendSignal(&util.Signal{ 
			Market: rt.market, 
			Side: "long",
			Reason: "Supertrend",
			Open: candle.Close * 1.01,
			TakeProfit: rt.takeProfitPrice,
			StopLoss: rt.stopLossPrice,
			UseTrailingStop: rt.useTrailingStop,
			Ratio: 1,
		})
		rt.openPosition("long", rt.initBalance, candle.Close, "Supertrend")
	}
	roi := util.CalcROI(rt.initBalance, rt.balance)
	util.Info(rt.tag, 
		fmt.Sprintf("balance: %.2f, total ROI: %.2f%%", rt.balance, roi * 100))
	winRate := float64(rt.takeProfitCount) / 
		float64(rt.takeProfitCount + rt.stopLossCount)
	util.Info(rt.tag, fmt.Sprintf("win rate: %.2f%%", winRate * 100))
	rt.prevSupertrend = supertrend
	rt.prevTrend = rt.trend
	rt.prevMainTrend = rt.mainTrend
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
	if len(candles) != rt.warmUpCandleNum {
		util.Error(rt.tag, "Error on getting warmup candles")
	}
	for _, candle := range candles {
		rt.prevSupertrend = rt.st.Update(candle)
		rt.prevTrend = rt.trend
		if candle.Close > rt.prevSupertrend {
			rt.trend = "bull"
		} else if candle.Close < rt.prevSupertrend {
			rt.trend = "bear"
		}
	}
	mainCandleStart := len(candles) - 1
	for ; mainCandleStart >= 0; mainCandleStart-- {
		candleTime := candles[mainCandleStart].GetTime()
		if candleTime.Unix() % int64(rt.mainRes) == 0 {
			break
		}
	}
	rt.mainCandle = candles[mainCandleStart]
	for i := mainCandleStart + 1; i < len(candles); i++ {
		rt.mainCandle.Update(candles[i])
	}
	candles = rt.getCandles(from, rt.mainRes)
	for _, candle := range candles {
		mainSupertrend := rt.mainST.Update(candle)
		rt.prevMainTrend = rt.mainTrend
		if candle.Close > mainSupertrend {
			rt.mainTrend = "bull"
		} else if candle.Close < mainSupertrend {
			rt.mainTrend = "bear"
		}
	}
}
func (rt *ResTrend) Start() {
	rt.warmUp(time.Now().Unix())
	candleChan := make(chan *util.Candle)
	go rt.ftx.SubCandle(rt.market, rt.res, candleChan);
	for candle := range candleChan {
		rt.genSignal(candle)
	}
}
