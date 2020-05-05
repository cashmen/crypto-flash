package indicator

import util "github.com/CheshireCatNick/crypto-flash/pkg/util"
import "math"

type Supertrend struct {
	Tag string
	period int
	multiplier float64
	atr *ATR
	prevAtr float64
	prevFinalUpperBand float64
	prevFinalLowerBand float64
	prevTrend string
	prevCandle *util.Candle
	prevPredictFinalUpperBand float64
	prevPredictFinalLowerBand float64
	prevPredictTrend string
}

func NewSupertrend(multiplier float64, period int) *Supertrend {
	return &Supertrend{
		Tag: "Supertrend",
		period: period,
		multiplier: multiplier,
		atr: NewATR(period),
		prevCandle: nil,
	}
}
func (st *Supertrend) CalculateSupertrend(candles []*util.Candle) []float64 {
	tst := NewSupertrend(st.multiplier, st.period)
	result := []float64{}
	for _, candle := range candles {
		result = append(result, tst.Update(candle))
	}
	return result;
}
func (st *Supertrend) Update(candle *util.Candle) float64 {
	atr := st.atr.Update(candle)
	basicUpperBand := candle.GetAvg() + st.multiplier * atr
	basicLowerBand := candle.GetAvg() - st.multiplier * atr
	var finalUpperBand, finalLowerBand, supertrend float64
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
	/* another version
	if candle.Close <= finalUpperBand {
		superTrend = finalUpperBand
	} else {
		superTrend = finalLowerBand
	}*/
	if candle.Close >= finalUpperBand {
		util.Info(st.Tag, util.Green("trend up"))
		supertrend = finalLowerBand
		st.prevTrend = "up"
		st.prevPredictTrend = "up"
	} else if candle.Close <= finalLowerBand {
		util.Info(st.Tag, util.Red("trend down"))
		supertrend = finalUpperBand
		st.prevTrend = "down"
		st.prevPredictTrend = "down"
	} else {
		// final lower band < close < final upper band
		// keep previous trend
		if (st.prevTrend == "up") {
			supertrend = finalLowerBand
		} else if (st.prevTrend == "down") {
			supertrend = finalUpperBand
		} else {
			supertrend = -1
		}
	}
	st.prevFinalUpperBand = finalUpperBand
	st.prevFinalLowerBand = finalLowerBand
	st.prevPredictFinalUpperBand = finalUpperBand
	st.prevPredictFinalLowerBand = finalLowerBand
	st.prevCandle = candle
	st.prevAtr = atr
	return supertrend
}
// predict works similar to update except it does not update internal state
// it predicts larger candle future supertrend by smaller candle
// should not use atr calculate from smaller candle to avoid bias
func (st *Supertrend) Predict(candle *util.Candle) float64 {
	atr := math.Max(st.atr.Predict(candle), st.prevAtr)
	basicUpperBand := candle.GetAvg() + st.multiplier * atr
	basicLowerBand := candle.GetAvg() - st.multiplier * atr
	var finalUpperBand, finalLowerBand, supertrend float64
	if st.prevCandle == nil {
		finalUpperBand = basicUpperBand
	} else if basicUpperBand < st.prevPredictFinalUpperBand || 
		st.prevCandle.Close > st.prevPredictFinalUpperBand {
		// price is falling or in up trend, adjust upperband
		finalUpperBand = basicUpperBand
	} else {
		// price is rising, maintain upperband
		finalUpperBand = st.prevPredictFinalUpperBand
	}
	if st.prevCandle == nil {
		finalLowerBand = basicLowerBand
	} else if basicLowerBand > st.prevPredictFinalLowerBand ||
		st.prevCandle.Close < st.prevPredictFinalLowerBand {
		// price is rising or in down trend, adjust lowerband
		finalLowerBand = basicLowerBand
	} else {
		// price is falling, maintain lowerband
		finalLowerBand = st.prevPredictFinalLowerBand
	}
	/* another version
	if candle.Close <= finalUpperBand {
		superTrend = finalUpperBand
	} else {
		superTrend = finalLowerBand
	}*/
	if candle.Close >= finalUpperBand {
		util.Info(st.Tag, util.Green("trend up"))
		supertrend = finalLowerBand
		st.prevPredictTrend = "up"
	} else if candle.Close <= finalLowerBand {
		util.Info(st.Tag, util.Red("trend down"))
		supertrend = finalUpperBand
		st.prevPredictTrend = "down"
	} else {
		// final lower band < close < final upper band
		// keep previous trend
		if (st.prevPredictTrend == "up") {
			supertrend = finalLowerBand
		} else if (st.prevPredictTrend == "down") {
			supertrend = finalUpperBand
		} else {
			supertrend = -1
		}
	}
	st.prevPredictFinalUpperBand = finalUpperBand
	st.prevPredictFinalLowerBand = finalLowerBand
	return supertrend
}
