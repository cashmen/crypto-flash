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
	"sort"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
)
type future struct {
	name string
	fundingRates []float64
	consCount int
	estApr float64
	size float64
	totalProfit float64
}
type FRArb struct {
	SignalProvider
	ftx *exchange.FTX
	// strategy config
	quarterContractName string
	updatePeriod int64
	futureNames []string
	leverage float64
	longTime int
	aprThreshold float64
	prevRateDays int64
	minAmount float64
	// data
	freeBalance float64
	futures map[string]*future
	startFutures []*future
	stopFutures []*future
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
		quarterContractName: "0626",
		updatePeriod: 15,
		futureNames: []string{
			"BTC", "ETH",
			
			"EOS", "LTC",
			"XTZ",
			"BSV", "BCH", "LINK", "ALT", "XRP",
			"BNB", "ATOM", "TRX", "ETC", "ALGO",
			"ADA", "SHIT", "MID", "MATIC", "EXCH",
			"HT", "TOMO", "DRGN", "TRYB", "XAUT",
			"OKB", "BTMX", "PRIV", "PAXG", "LEO",
			"DOGE", "USDT",
		},
		// perp and quarter have 1/2 pairPortion and leverage
		leverage: 5,
		// 5 consecutive hours of positive/negative funding rate
		longTime: 5 * 24,
		aprThreshold: 0.3,
		prevRateDays: 7,
		// minimum USD amount to start a pair (perp + quarter)
		minAmount: 10,
		// data
		futures: make(map[string]*future),
		freeBalance: 1000000,
	}
}
func (fra *FRArb) getFutureName(future string, isPerp bool) string {
	if isPerp {
		return future + "-PERP"
	} else {
		return future + "-" + fra.quarterContractName
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

func (fra *FRArb) genSignal(future *future) {
	fundingRates := future.fundingRates
	util.Info(fra.tag, future.name, 
		fmt.Sprintf("latestFundingRate: %f", fundingRates[0]))
	future.consCount = 1
	future.estApr = 1 + math.Abs(fundingRates[0])
	for i := 1; i < len(future.fundingRates); i++ {
		if fundingRates[i] * fundingRates[0] <= 0 {
			break
		}
		future.estApr *= 1 + math.Abs(fundingRates[i])
		future.consCount += 1
	}
	toAnnual := float64(365 * 24) / float64(future.consCount)
	future.estApr = (future.estApr - 1) * toAnnual * fra.leverage / 2
	util.Info(fra.tag, future.name, 
		fmt.Sprintf("estApr: %.2f%%", future.estApr * 100))
	if future.consCount >= fra.longTime && future.estApr >= fra.aprThreshold {
		if future.size == 0 {
			util.Info(fra.tag, "profitable: " + future.name)
			if fra.notifier != nil {
				fra.notifier.Broadcast(fra.tag, 
					"profitable: " + future.name + "\n" +
					fmt.Sprintf("estApr: %.2f%%", future.estApr * 100))
			}
			fra.startFutures = append(fra.startFutures, future)
		}
	} else {
		if future.size != 0 {
			util.Info(fra.tag, "not profitable: " + future.name)
			if fra.notifier != nil {
				fra.notifier.Broadcast(fra.tag, 
					"not profitable: " + future.name)
			}
			fra.stopFutures = append(fra.stopFutures, future)
		}
	}
}
func (fra* FRArb) sortApr() []string {
	type kv struct {
        k string
        v float64
    }
    var kvs []kv
    for name, future := range fra.futures {
        kvs = append(kvs, kv{name, future.estApr})
    }
    sort.Slice(kvs, func(i, j int) bool {
        return kvs[i].v > kvs[j].v
	})
	var names []string
    for _, kv := range kvs {
		names = append(names, kv.k)
	}
	return names
}
func (fra *FRArb) sendReport() {
	if fra.notifier == nil {
		return;
	}
	msg := "Report\n"
	runTime := time.Now().Sub(fra.startTime)
	d := util.FromTimeDuration(runTime)
	msg += "Runtime: " + d.String() + "\n\n"
	names := fra.sortApr()
	totalProfit := 0.0
	for _, name := range names {
		future := fra.futures[name]
		msg += "future: " + future.name + "\n"
		msg += fmt.Sprintf("estApr: %.2f%%\n", future.estApr * 100)
		msg += fmt.Sprintf("consCount: %d\n", future.consCount)
		msg += fmt.Sprintf("next funding rate: %f\n", future.fundingRates[0])
		msg += fmt.Sprintf("size: %f\n", future.size)
		msg += fmt.Sprintf("future profit: %f\n\n", future.totalProfit)
		totalProfit += future.totalProfit
	}
	msg += fmt.Sprintf("Total Profit: %.2f\n", totalProfit)
	balance := fra.initBalance + totalProfit
	roi := util.CalcROI(fra.initBalance, balance)
	msg += fmt.Sprintf("ROI: %.2f%%\n", roi * 100)
	ar := util.CalcAnnualFromROI(roi, runTime.Seconds())
	msg += fmt.Sprintf("Annualized Return: %.2f%%", ar * 100)
	fra.notifier.Broadcast(fra.tag, msg)
}
func (fra *FRArb) startPair(future *future, ratio float64) {
	/*
	perpSide := "long"
	quarterSide := "short"
	if future.size < 0 {
		// long pays short, short perp, long quarter
		perpSide = "short"
		quarterSide = "long"
	}
	// TODO: set stop loss
	fra.sendSignal(&util.Signal{ 
		Market: fra.getFutureName(future.name, true), 
		Side: perpSide,
		Reason: "Profitable",
		Ratio: ratio,
	})
	fra.sendSignal(&util.Signal{ 
		Market: fra.getFutureName(future.name, false), 
		Side: quarterSide,
		Reason: "Profitable",
		Ratio: ratio,
	})*/
	util.Info(fra.tag, fmt.Sprintf("start earning on %s, size %f",
		future.name, future.size))
	if fra.notifier != nil {
		fra.notifier.Broadcast(fra.tag, 
			fmt.Sprintf("start earning on %s, size %f",
			future.name, future.size))
	}
}
func (fra *FRArb) stopPair(future *future) {
	/*
	fra.sendSignal(&util.Signal{ 
		Market: fra.getFutureName(future.name, true), 
		Side: "close",
		Reason: "Not profitable",
	})
	fra.sendSignal(&util.Signal{ 
		Market: fra.getFutureName(future.name, false), 
		Side: "close",
		Reason: "Not profitable",
	})*/
	util.Info(fra.tag, fmt.Sprintf("stop earning on %s, size %f",
		future.name, future.size))
	if fra.notifier != nil {
		fra.notifier.Broadcast(fra.tag, 
			fmt.Sprintf("stop earning on %s, size %f",
			future.name, future.size))
	}
	pairPortion := math.Abs(future.size) / fra.leverage * 2
	fra.freeBalance += pairPortion
	future.size = 0
}
func (fra *FRArb) Start() {
	// get previous funding rate
	now := time.Now().Unix()
	end := now - now % (60 * 60)
	start := end - fra.prevRateDays * 24 * 60 * 60
	for _, name := range fra.futureNames {
		fra.futures[name] = &future{
			name: name,
		}
		fra.futures[name].fundingRates = 
			fra.ftx.GetFundingRates(start, end, fra.getFutureName(name, true))
	}
	for {
		now = time.Now().Unix()
		// TODO: check existing position every updatePeriod
		// one hour just passed, get funding rate of the previous hour
		if now % (60 * 60) == fra.updatePeriod {
			for name, future := range fra.futures {
				rates := fra.ftx.GetFundingRates(now - 60, now, 
					fra.getFutureName(name, true))
				future.fundingRates = 
					append([]float64{rates[0]}, 
					future.fundingRates[:24 * fra.prevRateDays - 1]...)
				// calculate profit if future has position
				future.totalProfit += future.size * future.fundingRates[0] * -1
				fra.genSignal(future)
			}
			for _, future := range fra.stopFutures {
				fra.stopPair(future)		
			}
			util.Info(fra.tag, fmt.Sprintf("free balance: %f, count: %d", 
				fra.freeBalance, len(fra.startFutures)))
			count := float64(len(fra.startFutures))
			if count > 0 && fra.freeBalance >= fra.minAmount * count {
				pairPortion := fra.freeBalance / count
				size := pairPortion / 2 * fra.leverage
				for _, future := range fra.startFutures {
					if future.fundingRates[0] > 0 {
						future.size = -size
					} else {
						future.size = size
					}
					fra.startPair(future, 1 / count / 2 * fra.leverage)
				}
				fra.freeBalance = 0
			}
			fra.startFutures = fra.startFutures[:0]
			fra.stopFutures = fra.stopFutures[:0]
			names := fra.sortApr()
			util.Info(fra.tag, "estApr Rank:")
			for _, name := range names {
				future := fra.futures[name]
				fmt.Printf("future: %s, estApr: %.2f%%, consCount: %d\n", 
				name, future.estApr * 100, future.consCount)
			}
		}
		// 8 hour just passed, generate report
		if now % (8 * 60 * 60) == fra.updatePeriod {
			fra.sendReport()
		}	
		timeToNextCycle := 
			fra.updatePeriod - time.Now().Unix() % fra.updatePeriod
		sleepDuration := util.Duration{Second: timeToNextCycle}
		time.Sleep(sleepDuration.GetTimeDuration())
	}
}


