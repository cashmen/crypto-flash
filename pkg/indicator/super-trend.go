package indicator

import util "github.com/CheshireCatNick/crypto-flash/pkg/util"
import "fmt"

type SuperTrend struct {
	period int
	multiplier float64
	atr *ATR
	prevFinalUpperBand float64
	prevFinalLowerBand float64
	prevTrend string
	prevCandle *util.Candle
}

func NewSuperTrend(multiplier float64, period int) *SuperTrend {
	return &SuperTrend{
		period: period,
		multiplier: multiplier,
		atr: NewATR(period),
		prevCandle: nil,
	}
}
func (st *SuperTrend) CalculateSuperTrend(candles []*util.Candle) []float64 {
	tst := NewSuperTrend(st.multiplier, st.period)
	result := []float64{}
	for _, candle := range candles {
		result = append(result, tst.Update(candle))
	}
	return result;
}
func (st *SuperTrend) Update(candle *util.Candle) float64 {
	atr := st.atr.Update(candle)
	basicUpperBand := candle.GetAvg() + st.multiplier * atr
	basicLowerBand := candle.GetAvg() - st.multiplier * atr
	var finalUpperBand, finalLowerBand, superTrend float64
	if st.prevCandle == nil {
		finalUpperBand = basicUpperBand
	} else if basicUpperBand < st.prevFinalUpperBand || 
		st.prevCandle.Close > st.prevFinalUpperBand {
		// price is falling or in up trend, adjust upperband
		finalUpperBand = basicUpperBand
	} else {
		// price is rising, maintain upperband
		finalUpperBand = st.prevFinalUpperBand
	}
	if st.prevCandle == nil {
		finalLowerBand = basicLowerBand
	} else if basicLowerBand > st.prevFinalLowerBand ||
		st.prevCandle.Close < st.prevFinalLowerBand {
		// price is rising or in down trend, adjust lowerband
		finalLowerBand = basicLowerBand
	} else {
		// price is falling, maintain lowerband
		finalLowerBand = st.prevFinalLowerBand
	}
	/*
	if candles[i].Close <= finalUpperBand {
		superTrend = finalUpperBand
	} else {
		superTrend = finalLowerBand
	}*/
	if candle.Close >= finalUpperBand {
		fmt.Println("up")
		superTrend = finalLowerBand
		st.prevTrend = "up"
	} else if candle.Close <= finalLowerBand {
		fmt.Println("down")
		superTrend = finalUpperBand
		st.prevTrend = "down"
	} else {
		// final lower band < close < final upper band
		// keep previous trend
		if (st.prevTrend == "up") {
			superTrend = finalLowerBand
		} else {
			superTrend = finalUpperBand
		}
	}
	st.prevFinalUpperBand = finalUpperBand
	st.prevFinalLowerBand = finalLowerBand
	st.prevCandle = candle
	return superTrend
}