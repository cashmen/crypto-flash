package util

type Candle struct {
	close     float64
	high      float64
	low       float64
	open      float64
	startTime string
	volume    float64
}

func NewCandle(c, h, l, o, v float64, st string) *Candle {
	var candle Candle = Candle{ 
		close: c, 
		high: h,
		low: l,
		open: o,
		volume: v,
		startTime: st,
	}
	return &candle;
}

func (candle *Candle) GetClose() float64 {
	return candle.close;
}
func (candle *Candle) GetHigh() float64 {
	return candle.high;
}
func (candle *Candle) GetLow() float64 {
	return candle.low;
}
func (candle *Candle) GetOpen() float64 {
	return candle.open;
}
func (candle *Candle) GetAvg() float64 {
	return (candle.high + candle.low) / 2
}