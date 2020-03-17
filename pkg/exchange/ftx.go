/* 
// Ftx implements exchange API for FTX exchange.
// Input: real exchange, trader
// Output: real exchange, signal provider or indicator
// TODO: 
// 1. getHistoryCandles: if candles >= 5000, request many times and concat result
*/
package exchange

import (
	"encoding/json"
	"net/http"
	"fmt"
	"log"
	"time"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
)

const (
	apiEndPoint string = "https://ftx.com/api"
	tag = "FTX"
)

type FTX struct {
	// save all candles data from different resolutions and markets
	candleData map[string][]*util.Candle
	candleSubs map[string][]chan<- *util.Candle
}

func NewFTX() *FTX {
	var ftx FTX
	ftx.candleData = make(map[string][]*util.Candle)
	ftx.candleSubs = make(map[string][]chan<- *util.Candle)
	return &ftx
}
func (ftx *FTX) GetHistoryCandles(market string, resolution int,
	startTime int64, endTime int64) []*util.Candle {
	type candleResp struct {
		Close     float64
		High      float64
		Low       float64
		Open      float64
		StartTime string
		Volume    float64
	}
	type historyResp struct {
		Success bool
		Result  []candleResp
	}
	req := fmt.Sprintf(
		"%s/markets/%s/candles?" + 
		"resolution=%d&start_time=%d&end_time=%d&limit=5000",
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
	var candles []*util.Candle
	for _, c := range resObj.Result {
		candles = append(candles, util.NewCandle(
			c.Open, c.High, c.Low, c.Close, c.Volume, c.StartTime))
	}
	return candles
}
func sleepToNextCandle(resolution int64) {
	timeToNextCandle := resolution - time.Now().Unix() % resolution
	sleepDuration := util.Duration{Second: timeToNextCandle + 1}
	time.Sleep(sleepDuration.GetTimeDuration())
}
// resolution can be 15, 60, 300, 900, 3600, 14400, 86400
func (ftx *FTX) SubCandle(
		market string, resolution int, c chan<- *util.Candle) {
	dataID := fmt.Sprintf("%s-%d", market, resolution)
	if _, exist := ftx.candleData[dataID]; exist {
		// someone already sub this data
		ftx.candleSubs[dataID] = append(ftx.candleSubs[dataID], c)
		return;
	}
	ftx.candleData[dataID] = []*util.Candle{}
	ftx.candleSubs[dataID] = []chan<- *util.Candle{}
	ftx.candleSubs[dataID] = append(ftx.candleSubs[dataID], c)
	resolution64 := int64(resolution)
	// sleep to the next candle
	sleepToNextCandle(resolution64)
	for {
		now := time.Now().Unix()
		startTime := now - resolution64 * 2 + 1
		endTime := now - resolution64
		candles := ftx.GetHistoryCandles(
			"BTC-PERP", resolution, startTime, endTime)
		for _, c := range ftx.candleSubs[dataID] {
			c<-candles[0]
		}
		ftx.candleData[dataID] = append(ftx.candleData[dataID], candles...)
		sleepToNextCandle(resolution64)
	}
}