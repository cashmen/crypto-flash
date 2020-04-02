package util

import "fmt"

type Position struct {
	Market string
	Side string
	Size float64
	OpenPrice float64
}

const tag = "Position"

func NewPosition(side string, size float64, openPrice float64) *Position {
	return &Position{ Side: side, Size: size, OpenPrice: openPrice }
}
func (pos *Position) Close(closePrice float64) float64 {
	roi := (closePrice - pos.OpenPrice) / pos.OpenPrice
	if pos.Side == "short" {
		roi *= -1
	}
	roiStr := PF64(roi * 100)
	if roi > 0 {
		roiStr = Green(roiStr)
	} else {
		roiStr = Red(roiStr)
	}
	Info(tag, fmt.Sprintf(
		"close %s, open price: %.2f, current price: %.2f, ROI: %s", 
		pos.Side, pos.OpenPrice, closePrice, roiStr))
	return roi
}
func (pos *Position) String() string {
	return fmt.Sprintf("Market: %s, Side: %s, Size: %f, OpenPrice: %f",
		pos.Market, pos.Side, pos.Size, pos.OpenPrice)
}