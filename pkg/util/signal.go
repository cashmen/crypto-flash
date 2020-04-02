package util

type Signal struct {
	Market string
	Side string
	Reason string
	Open float64
	TakeProfit float64
	StopLoss float64
	UseTrailingStop bool
}