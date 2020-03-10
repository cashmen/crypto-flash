package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"

	cryptoflash "github.com/CheshireCatNick/crypto-flash/pkg/crypto-flash"
)

type config struct {
	Key        string
	Secret     string
	SubAccount string
}

type candle struct {
	Close     float64
	High      float64
	Low       float64
	Open      float64
	StartTime string
	Volume    float64
}

type duration struct {
	year   int64
	month  int64
	day    int64
	hour   int64
	minute int64
	second int64
}

const (
	apiEndPoint string = "https://ftx.com/api"
)

func loadConfig(fileName string) config {
	var c config
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bytes, &c)
	return c
}

func calculateTrueRanges(candles []candle) []float64 {
	var result []float64
	for i := 0; i < len(candles); i++ {
		if i == 0 {
			result = append(result, candles[0].High-candles[0].Low)
			continue
		}
		a := candles[i].High - candles[i].Low
		b := math.Abs(candles[i].High - candles[i-1].Close)
		c := math.Abs(candles[i].Low - candles[i-1].Close)
		result = append(result, math.Max(math.Max(a, b), c))
	}
	return result
}
func calculateSMA(period int, arr []float64) []float64 {
	var result []float64
	sum := 0.0
	windowSize := 0
	for i := 0; i < len(arr); i++ {
		sum += arr[i]
		if windowSize < period {
			windowSize++
		}
		result = append(result, sum/float64(windowSize))
		if i >= period {
			sum -= arr[i-period]
		}
	}
	return result
}
func calculateRMA(period int, arr []float64) []float64 {
	var result []float64
	windowSize := 1
	prevRMA := arr[0]
	result = append(result, arr[0])
	for i := 1; i < len(arr); i++ {
		if windowSize < period {
			windowSize++
		}
		prevRMA = (prevRMA*(float64(windowSize)-1) + arr[i]) / float64(windowSize)
		result = append(result, prevRMA)
	}
	return result
}
func calculateATR(period int, candles []candle) []float64 {
	return calculateRMA(period, calculateTrueRanges(candles))
}
func calculateSuperTrend(multiplier int, period int, candles []candle) []float64 {
	atrs := calculateATR(period, candles)
	var result []float64

	for i := 0; i < len(atrs); i++ {
		//fmt.Println(atrs[i])
	}
	fmt.Println(len(candles))
	return result
}
func getHistoryCandle(marketName string, resolution int,
	startTime int64, endTime int64) []candle {
	type historyResp struct {
		Success bool
		Result  []candle
	}
	req := fmt.Sprintf(
		"%s/markets/%s/candles?resolution=%d&start_time=%d&end_time=%d&limit=5000",
		apiEndPoint, marketName, resolution, startTime, endTime)
	fmt.Println(req)
	res, err := http.Get(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	var resObj historyResp
	if decoder.Decode(&resObj) != nil {
		log.Fatal(err)
	}
	fmt.Println(resObj)
	return resObj.Result
}
func getDuration(d duration) time.Duration {
	sToNano := int64(1000000000)
	mToNano := sToNano * 60
	hToNano := mToNano * 60
	dToNano := hToNano * 24
	monthToNano := dToNano * 30
	yToNano := monthToNano * 12
	return time.Duration(-1 * (d.year*yToNano + d.month*monthToNano +
		d.day*dToNano + d.hour*hToNano + d.minute*mToNano + d.second*sToNano))
}
func main() {
	config := loadConfig("config.json")
	fmt.Println(config)

	trader := cryptoflash.NewTrader()
	trader.Run()

	res, err := http.Get("https://ftx.com/api/markets")
	if err != nil {
		log.Fatal(err)
	}
	//robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s", robots)
	endTime := time.Now()
	fmt.Println(endTime)
	var d duration
	d.day = 365
	fmt.Println(d)
	duration := getDuration(d)
	/*
		duration, err := time.ParseDuration("-10h")
		if err != nil {
			log.Fatal(err)
		}*/
	startTime := endTime.Add(duration)
	//startTime = startTime.AddDate(0, -1, 0)
	candles := getHistoryCandle("BTC-PERP", 4*3600, startTime.Unix(), endTime.Unix())
	superTrends := calculateSuperTrend(3, 10, candles)
	fmt.Println(superTrends)

}
