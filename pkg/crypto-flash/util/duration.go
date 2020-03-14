package util

import "time"
type Duration struct {
	Year   int64
	Month  int64
	Day    int64
	Hour   int64
	Minute int64
	Second int64
}

// convert my duration to time.Duration
func (d *Duration) GetDuration() time.Duration {
	sToNano := int64(1000000000)
	mToNano := sToNano * 60
	hToNano := mToNano * 60
	dToNano := hToNano * 24
	monthToNano := dToNano * 30
	yToNano := monthToNano * 12
	return time.Duration(d.Year*yToNano + d.Month*monthToNano +
		d.Day*dToNano + d.Hour*hToNano + d.Minute*mToNano + d.Second*sToNano)
}