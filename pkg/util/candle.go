package util

import "fmt"

type Candle struct {
	Close     float64
	High      float64
	Low       float64
	Open      float64
	StartTime string
	Volume    float64
}

func NewCandle(o, l, h, c, v float64, st string) *Candle {
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
func (candle *Candle) String() string {
	var color func(string, ...interface{}) string
	if candle.Close > candle.Open {
		color = Green
	} else {
		color = Red
	}
	return color(fmt.Sprintf("o: %.2f, l: %.2f, h: %.2f, c: %.2f, time: %s", 
		candle.Open, candle.Low, candle.High, candle.Close, candle.StartTime))
}