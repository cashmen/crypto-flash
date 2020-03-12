/*
// The main program of crypto flash.
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"

	//cryptoflash "github.com/CheshireCatNick/crypto-flash/pkg/crypto-flash"
)

type config struct {
	Key        string
	Secret     string
	SubAccount string
	Channel_Secret string
	Channel_Access_Token string
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
func calculateSuperTrends(multiplier float64, period int, candles []candle) []float64 {
	atrs := calculateATR(period, candles)
	var result []float64
	var (
		basicUpperBand float64
		basicLowerBand float64
		finalUpperBand float64
		finalLowerBand float64
		prevFinalUpperBand float64
		prevFinalLowerBand float64
		superTrend float64
		prevTrend string
	)
	prevTrend = "unknown"
	for i := 0; i < len(candles); i++ {
		basicUpperBand = 
			(candles[i].High + candles[i].Low) / 2 + multiplier * atrs[i]
		basicLowerBand = 
			(candles[i].High + candles[i].Low) / 2 - multiplier * atrs[i]
		if i == 0 {
			finalUpperBand = basicUpperBand
		} else if basicUpperBand < prevFinalUpperBand || 
			candles[i - 1].Close > prevFinalUpperBand {
			// price is falling or in up trend, adjust upperband
			finalUpperBand = basicUpperBand
		} else {
			// price is rising, maintain upperband
			finalUpperBand = prevFinalUpperBand
		}
		if i == 0 {
			finalLowerBand = basicLowerBand
		} else if basicLowerBand > prevFinalLowerBand ||
			candles[i - 1].Close < prevFinalLowerBand {
			// price is rising or in down trend, adjust lowerband
			finalLowerBand = basicLowerBand
		} else {
			// price is falling, maintain lowerband
			finalLowerBand = prevFinalLowerBand
		}
		/*
		if candles[i].Close <= finalUpperBand {
			superTrend = finalUpperBand
		} else {
			superTrend = finalLowerBand
		}*/
		if candles[i].Close >= finalUpperBand {
			fmt.Printf("up %d\n", i)
			superTrend = finalLowerBand
			prevTrend = "up"
		} else if candles[i].Close <= finalLowerBand {
			fmt.Printf("down %d\n", i)
			superTrend = finalUpperBand
			prevTrend = "down"
		} else {
			// final lower band < close < final upper band
			// keep previous trend
			if (prevTrend == "up") {
				superTrend = finalLowerBand
			} else {
				superTrend = finalUpperBand
			}
		}
		fmt.Println(superTrend)
		result = append(result, superTrend)
		prevFinalUpperBand = finalUpperBand
		prevFinalLowerBand = finalLowerBand
	}
	//fmt.Println(len(candles))
	return result
}
func getHistoryCandles(market string, resolution int,
	startTime int64, endTime int64) []candle {
	type historyResp struct {
		Success bool
		Result  []candle
	}
	req := fmt.Sprintf(
		"%s/markets/%s/candles?resolution=%d&start_time=%d&end_time=%d&limit=5000",
		apiEndPoint, market, resolution, startTime, endTime)
	//fmt.Println(req)
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
	//fmt.Println(resObj)
	return resObj.Result
}
// convert my duration to time.Duration
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
func closePosition(side string, openPrice, closePrice float64) float64 {
	ROI := (closePrice - openPrice) / openPrice
	if side == "short" {
		ROI *= -1
	}
	fmt.Printf("close %s, open price: %f, current price: %f\n", side, openPrice, closePrice)
	fmt.Printf("ROI %f\n", ROI)
	return ROI
}
func backtest(market string, resolution int, startTime, endTime int64) {
	candles := getHistoryCandles(market, resolution, startTime, endTime)
	superTrends := calculateSuperTrends(3, 10, candles)
	initUSD := 1000000.0
	currentUSD := initUSD
	currentState := "no pos"
	prevState := ""
	takeProfit := 300.0
	stopLoss := 100.0
	var currentPos float64 = -1
	fmt.Println("start backtesting")
	for i := 0; i < len(candles) - 1; i++ {
		fmt.Printf("close %f, st: %f\n", candles[i].Close, superTrends[i])
		// take profit or stop loss
		if currentState == "long" {
			if candles[i].High - currentPos >= takeProfit {
				ROI := closePosition("long", currentPos, currentPos + takeProfit)
				currentUSD *= (1 + ROI)
				prevState = currentState
				currentState = "no pos"
			} else if (currentPos - candles[i].Low >= stopLoss) {
				ROI := closePosition("long", currentPos, currentPos - stopLoss)
				currentUSD *= (1 + ROI)
				prevState = currentState
				currentState = "no pos"
			}
		} else if currentState == "short" {
			if candles[i].High - currentPos >= stopLoss {
				ROI := closePosition("short", currentPos, currentPos + stopLoss)
				currentUSD *= (1 + ROI)
				prevState = currentState
				currentState = "no pos"
			} else if (currentPos - candles[i].Low >= takeProfit) {
				ROI := closePosition("short", currentPos, currentPos - takeProfit)
				currentUSD *= (1 + ROI)
				prevState = currentState
				currentState = "no pos"
			}
		}
		if currentState != "short" && candles[i].Close < superTrends[i] &&
			prevState != "short" {
			if currentState == "long" {
				// close long position
				ROI := closePosition("long", currentPos, candles[i + 1].Open)
				currentUSD *= (1 + ROI)
			}
			fmt.Println("start short")
			currentState = "short"
			currentPos = candles[i + 1].Open
		} else if currentState != "long" && candles[i].Close > superTrends[i] &&
			prevState != "long" {
			if currentState == "short" {
				// close short position
				ROI := closePosition("short", currentPos, candles[i + 1].Open)
				currentUSD *= (1 + ROI)
			}
			fmt.Println("start long")
			currentState = "long"
			currentPos = candles[i + 1].Open
		}
	}
	ROI := (currentUSD - initUSD) / initUSD
	fmt.Printf("final balance: %f, total ROI: %f\n", currentUSD, ROI)
}
func main() {
	//config := loadConfig("config.json")
	endTime := time.Now()
	var d duration
	d.day = 30
	duration := getDuration(d)
	startTime := endTime.Add(duration)
	backtest("BTC-PERP", 1 * 3600, startTime.Unix(), endTime.Unix())
	//fmt.Println(superTrends)
	// test line bot function
	//notifier := cryptoflash.NewNotifier(config.Channel_Secret, config.Channel_Access_Token)
	//notifier.Broadcast("test")
}
