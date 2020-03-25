package util

import "time"
import "fmt"

type Duration struct {
	Year   int64
	Month  int64
	Day    int64
	Hour   int64
	Minute int64
	Second int64
}
const (
	sToNano = int64(1000000000)
	mToNano = sToNano * 60
	hToNano = mToNano * 60
	dToNano = hToNano * 24
	monthToNano = dToNano * 30
	yToNano = monthToNano * 12
)
// convert my duration to time.Duration
func (d *Duration) GetTimeDuration() time.Duration {
	return time.Duration(d.Year * yToNano + d.Month * monthToNano +
		d.Day * dToNano + d.Hour * hToNano + d.Minute * mToNano + 
		d.Second * sToNano)
}
func FromTimeDuration(td time.Duration) *Duration {
	nano := td.Nanoseconds()
	d := Duration{}
	d.Year = nano / yToNano
	nano = nano % yToNano
	d.Month = nano / monthToNano
	nano = nano % monthToNano
	d.Day = nano / dToNano
	nano = nano % dToNano
	d.Hour = nano / hToNano
	nano = nano % hToNano
	d.Minute = nano / mToNano
	nano = nano % mToNano
	d.Second = nano / sToNano
	return &d
}
func (d *Duration) String() string {
	return fmt.Sprintf("%dd%dh%dm%ds", d.Day, d.Hour, d.Minute, d.Second)
}