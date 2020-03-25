package util

func CalcROI(initBalance, balance float64) float64 {
	return (balance - initBalance) / initBalance
}
// time is in second
func CalcAnnual(initBalance, balance float64, time float64) float64 {
	return CalcROI(initBalance, balance) * (86400 * 365) / time
}
func CalcAnnualFromROI(roi float64, time float64) float64 {
	return roi * (86400 * 365) / time
}