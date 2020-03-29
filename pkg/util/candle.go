package util

import "fmt"
import "time"

type Candle struct {
	Close     float64
	High      float64
	Low       float64
	Open      float64
	StartTime string
	Volume    float64
}
func NewCandle(o, h, l, c, v float64, st string) *Candle {
	loc, _ := time.LoadLocation("Asia/Taipei")
	candleTime, _ := time.Parse(time.RFC3339, st)
	return &Candle{ 
		Close: c, 
		High: h,
		Low: l,
		Open: o,
		Volume: v,
		StartTime: candleTime.In(loc).String(),
	}
}
func (candle *Candle) Copy() *Candle {
	return &Candle{ 
		Close: candle.Close, 
		High: candle.High,
		Low: candle.Low,
		Open: candle.Open,
		Volume: candle.Volume,
		StartTime: candle.StartTime,
	}
}
func (candle *Candle) GetAvg() float64 {
	return (candle.High + candle.Low) / 2
}
func (candle *Candle) Update(smallCandle *Candle) {
	if candle.High < smallCandle.High {
		candle.High = smallCandle.High
	}
	if candle.Low > smallCandle.Low {
		candle.Low = smallCandle.Low
	}
	candle.Close = smallCandle.Close
}
func (candle *Candle) GetTime() time.Time {
	candleTime, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", candle.StartTime)
	return candleTime
}
func (candle *Candle) String() string {
	var color func(string, ...interface{}) string
	if candle.Close > candle.Open {
		color = Green
	} else {
		color = Red
	}
	return color(fmt.Sprintf("o: %.2f, h: %.2f, l: %.2f, c: %.2f, time: %s", 
		candle.Open, candle.High, candle.Low, candle.Close, candle.StartTime))
}