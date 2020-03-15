package util

import "fmt"
type Position struct {
	Side string
	Size float64
	OpenPrice float64
}

func NewPosition(side string, size float64, openPrice float64) *Position {
	return &Position{ Side: side, Size: size, OpenPrice: openPrice }
}
func (pos *Position) Close(closePrice float64) float64 {
	ROI := (closePrice - pos.OpenPrice) / pos.OpenPrice
	if pos.Side == "short" {
		ROI *= -1
	}
	fmt.Printf("close %s, open price: %f, current price: %f\n", 
		pos.Side, pos.OpenPrice, closePrice)
	fmt.Printf("ROI %f\n", ROI)
	return ROI
}