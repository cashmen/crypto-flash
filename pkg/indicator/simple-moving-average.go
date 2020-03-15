package indicator

type SMA struct {
	data []float64
	period int
	windowSize int
	sum float64
}

func NewSMA(period int) *SMA {
	return &SMA{ period: period }
}
func (sma *SMA) CalculateSMA(arr []float64) []float64 {
	tsma := NewSMA(sma.period)
	result := []float64{}
	for _, n := range arr {
		result = append(result, tsma.Update(n))
	}
	return result
}
func (sma *SMA) Update(val float64) float64 {
	sma.data = append(sma.data, val)
	if sma.windowSize < sma.period {
		sma.windowSize++
	}
	l := len(sma.data)
	if (l > sma.windowSize) {
		sma.sum -= sma.data[l - sma.windowSize - 1]
	}
	sma.sum += val
	result := sma.sum / float64(sma.windowSize)
	return result
}