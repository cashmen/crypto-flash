package indicator

type RMA struct {
	period int
	windowSize int
	prevRMA float64
}

func NewRMA(period int) *RMA {
	return &RMA{ period: period }
}
func (rma *RMA) CalculateRMA(arr []float64) []float64 {
	trma := NewRMA(rma.period)
	result := []float64{}
	for _, n := range arr {
		result = append(result, trma.Update(n))
	}
	return result;
}
func (rma *RMA) Update(val float64) float64 {
	if rma.windowSize < rma.period {
		rma.windowSize++
	}
	if (rma.windowSize == 1) {
		rma.prevRMA = val;
	} else {
		rma.prevRMA = (rma.prevRMA * (float64(rma.windowSize) - 1) + val) / 
			float64(rma.windowSize)
	}
	return rma.prevRMA
}
func (rma *RMA) Predict(val float64) float64 {
	windowSize := rma.windowSize
	if windowSize < rma.period {
		windowSize++
	}
	if (windowSize == 1) {
		return val
	} else {
		return (rma.prevRMA * (float64(windowSize) - 1) + val) / 
			float64(windowSize)
	}
}