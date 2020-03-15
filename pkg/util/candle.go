package util

type Candle struct {
	Close     float64
	High      float64
	Low       float64
	Open      float64
	StartTime string
	Volume    float64
}

func NewCandle(c, h, l, o, v float64, st string) *Candle {
	return &Candle{ 
		Close: c, 
		High: h,
		Low: l,
		Open: o,
		Volume: v,
		StartTime: st,
	}
}

func (candle *Candle) GetAvg() float64 {
	return (candle.High + candle.Low) / 2
}