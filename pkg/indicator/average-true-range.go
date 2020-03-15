package indicator

import util "github.com/CheshireCatNick/crypto-flash/pkg/util"
import "math"

type ATR struct {
	period int
	prevCandle *util.Candle
	rma *RMA
}

func NewATR(period int) *ATR{
	return &ATR{
		period: period,
		prevCandle: nil,
		rma: NewRMA(period),
	}
}
func (atr *ATR) CalculateATR(candles []*util.Candle) []float64 {
	tatr := NewATR(atr.period)
	result := []float64{}
	for _, candle := range candles {
		result = append(result, tatr.Update(candle))
	}
	return result;
}
func (atr *ATR) updateTR(candle *util.Candle) float64 {
	if atr.prevCandle == nil {
		atr.prevCandle = candle
		return candle.High - candle.Low
	}
	a := candle.High - candle.Low
	b := math.Abs(candle.High - atr.prevCandle.Close)
	c := math.Abs(candle.Low - atr.prevCandle.Close)
	atr.prevCandle = candle
	return math.Max(math.Max(a, b), c)
}
func (atr *ATR) Update(candle *util.Candle) float64 {
	return atr.rma.Update(atr.updateTR(candle))
}