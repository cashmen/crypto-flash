/* 
// Signal provider provide trade signals. A trade signal should include 
// informations such as buy/sell, stop-loss/take-profit prices, confidence, etc.
// Signal provider is an implementation of a strategy.
// Input: market data from exchanges or indicators
// Output: trader or notifier
// TODO: auto parameter adjustment
*/
package character

import (
	"fmt"
	"time"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	chart "github.com/wcharczuk/go-chart"
	drawing "github.com/wcharczuk/go-chart/drawing"
	"os"
)

type SignalProvider struct {
	tag string
	startTime time.Time
	position *util.Position
	initBalance float64
	balance float64
	notifier *Notifier
	signalChan chan<- *util.Signal
	chans []chan<- *util.Signal
	stopLossCount int
	takeProfitCount int
	profits []float64
}
func (sp *SignalProvider) showChart() {
	var ticks []float64
	var zeroes []float64
	for i := 1; i <= len(sp.profits); i++ {
		ticks = append(ticks, float64(i))
		zeroes = append(zeroes, 0)
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name: "Profit",
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("00897B"),
					FillColor: drawing.ColorFromHex("00897B").WithAlpha(60),
				},
				XValues: ticks,
				YValues: sp.profits,
			},
			chart.ContinuousSeries{
				Name: "Zero Line",
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("C62828"),
				},
				XValues: ticks,
				YValues: zeroes,
			},
		},
	}
	file, err := os.Create("backtest.png")
	err = graph.Render(chart.PNG, file)
	if err != nil {
		util.Error(err.Error())
	}
}
func (sp *SignalProvider) notifyROI() {
	if sp.notifier == nil {
		return;
	}
	roi := util.CalcROI(sp.initBalance, sp.balance)
	winRate := float64(sp.takeProfitCount) / 
		float64(sp.takeProfitCount + sp.stopLossCount)
	msg := "Report\n"
	runTime := time.Now().Sub(sp.startTime)
	d := util.FromTimeDuration(runTime)
	msg += "Runtime: " + d.String() + "\n"
	msg += fmt.Sprintf("Init Balance: %.2f\n", sp.initBalance)
	msg += fmt.Sprintf("Balance: %.2f\n", sp.balance)
	msg += fmt.Sprintf("Win Rate: %.2f%%\n", winRate * 100)
	msg += fmt.Sprintf("ROI: %.2f%%\n", roi * 100)
	ar := util.CalcAnnualFromROI(roi, runTime.Seconds())
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
func (sp *SignalProvider) closePosition(price float64, reason string) float64 {
	roi := sp.position.Close(price)
	sp.balance *= 1 + roi
	sp.profits = append(sp.profits, sp.balance - sp.initBalance)
	logMsg := fmt.Sprintf("close %s @ %.2f due to %s, ROI: %.2f%%", 
		sp.position.Side, price, reason, roi * 100)
	if roi > 0 { 
		util.Info(sp.tag, util.Green(logMsg))
		sp.takeProfitCount++
	} else {
		util.Info(sp.tag, util.Red(logMsg))
		sp.stopLossCount++
	}
	sp.notifyClosePosition(price, roi, reason)
	sp.position = nil
	return roi
}
func (sp *SignalProvider) openPosition(
		side string, size, price float64, reason string) {
	sp.position = util.NewPosition(side, size, price)
	logMsg := fmt.Sprintf("start %s @ %.2f due to %s", side, price, reason)
	if side == "long" {
		util.Info(sp.tag, util.Green(logMsg))
	} else {
		util.Info(sp.tag, util.Red(logMsg))
	}
	sp.notifyOpenPosition(reason)
}
func (sp *SignalProvider) sendSignal(s *util.Signal) {
	for _, c := range sp.chans {
		c<-s
	}
}
func (sp *SignalProvider) SubSignal(signalChan chan<- *util.Signal) {
	sp.chans = append(sp.chans, signalChan)
}
