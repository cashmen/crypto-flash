package util

import color "github.com/fatih/color"
import "fmt"
import "time"

const debug = true

var (
	Red = color.RedString
	Blue = color.HiBlueString
	Green = color.GreenString
	Yellow = color.YellowString
	Cyan = color.CyanString
)

func PrintRed(s string) {
	fmt.Println(Red(s))
}
func PrintGreen(s string) {
	fmt.Println(Green(s))
}
// parse
func PF64(n float64) string {
	return fmt.Sprintf("%.2f", n)
}
func PI(n int) string {
	return fmt.Sprintf("%d", n)
}
func PI64(n int64) string {
	return fmt.Sprintf("%d", n)
}
func print(color string, s ...string) {
	if !debug {
		return
	}
	tag := ""
	switch color {
	case "red":
		tag = fmt.Sprintf("%s", Red(s[0]))
	case "blue":
		tag = fmt.Sprintf("%s", Blue(s[0]))
	case "green":
		tag = fmt.Sprintf("%s", Green(s[0]))
	case "yello":
		tag = fmt.Sprintf("%s", Yellow(s[0]))
	}
	loc, _ := time.LoadLocation("Asia/Taipei")
	result := fmt.Sprintf("[%s] [%s]", 
		tag, Cyan(time.Now().In(loc).Format("1/2 15:04:05")))
	for i := 1; i < len(s); i++ {
		result += " " + s[i]
	}
	fmt.Println(result)
}
func Warning(s ...string) {
	print("yellow", s...)
}
func Info(s ...string) {
	print("blue", s...)
}
func Success(s ...string) {
	print("green", s...)
}
func Error(s ...string) {
	print("red", s...)
	//panic(s)
}