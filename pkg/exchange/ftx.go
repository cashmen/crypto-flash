/* 
// Ftx implements exchange API for FTX exchange.
// Input: real exchange, trader
// Output: real exchange, signal provider or indicator
// TODO: 
// 1. getHistoryCandles: if candles >= 5000, request many times and concat result
// 3. getPosition
// 4. makeOrder
// 5. exchange interface
*/
package exchange

import (
	"fmt"
	"time"
	"net/http"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
)

const (
	host string = "https://ftx.com"
	marketAPI string = "/api/markets"
	walletAPI string = "/api/wallet/balances"
	orderAPI string = "/api/orders"
	triggerOrderAPI string = "/api/conditional_orders"
)

type FTX struct {
	tag string
	key string
	subAccount string
	secret string
	// save all candles data from different resolutions and markets
	candleData map[string][]*util.Candle
	candleSubs map[string][]chan<- *util.Candle
	restClient *util.RestClient
}

func NewFTX(key, secret, subAccount string) *FTX {
	return &FTX{
		key: key,
		secret: secret,
		subAccount: subAccount,
		tag: "FTX",
		candleData: make(map[string][]*util.Candle),
		candleSubs: make(map[string][]chan<- *util.Candle),
		restClient: util.NewRestClient(),
	}
}
// depth 20 ~ 100
func (ftx *FTX) GetOrderbook(market string, depth int) *util.Orderbook {
	type orderbookRes struct {
		Asks [][2]float64
		Bids [][2]float64
	}
	type res struct {
		Success bool
		Result orderbookRes
	}
	url := host + marketAPI + 
		fmt.Sprintf("/%s/orderbook?depth=%d", market, depth)
	var resObj res
	ftx.restClient.Get(url, nil, &resObj)
	orderbook := &util.Orderbook{}
	for _, row := range resObj.Result.Asks {
		orderbook.Add("ask", row[0], row[1])
	}
	for _, row := range resObj.Result.Bids {
		orderbook.Add("bid", row[0], row[1])
	}
	return orderbook
}
func (ftx *FTX) GetHistoryCandles(market string, resolution int,
	startTime int64, endTime int64) []*util.Candle {
	type candleRes struct {
		Close     float64
		High      float64
		Low       float64
		Open      float64
		StartTime string
		Volume    float64
	}
	type historyRes struct {
		Success bool
		Result  []candleRes
	}
	url := host + marketAPI + fmt.Sprintf(
		"/%s/candles?resolution=%d&start_time=%d&end_time=%d&limit=5000",
		market, resolution, startTime, endTime)
	var resObj historyRes
	ftx.restClient.Get(url, nil, &resObj)
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
func (ftx *FTX) genAuthHeader(method, path, body string) *http.Header {
	header := http.Header(make(map[string][]string))
	header.Add("FTX-KEY", ftx.key)
	ts := fmt.Sprintf("%d", time.Now().UnixNano() / 1000000)
	header.Add("FTX-TS", ts)
	payload := ts + method + path + body
	signature := util.HMac(payload, ftx.secret)
	header.Add("FTX-SIGN", signature)
	header.Add("FTX-SUBACCOUNT", ftx.subAccount)
	return &header
}
func (ftx *FTX) GetWallet() *util.Wallet {
	type coin struct {
		Coin string
		Free float64
		Total float64
	}
	type res struct {
		Success bool
		Result []coin
	}
	url := host + walletAPI
	header := ftx.genAuthHeader("GET", walletAPI, "")
	var resObj res
	ftx.restClient.Get(url, header, &resObj)
	wallet := util.NewWallet()
	for _, coin := range resObj.Result {
		wallet.Increase(coin.Coin, coin.Free)
	}
	return wallet
}
func (ftx *FTX) MakeOrder(order *util.Order) int64 {
	type result struct {
		CreatedAt string
		FilledSize float64
		Future string
		Id int64
		Market string
		Price float64
		RemainSize float64
		Side string
		Size float64
		Status string
		Type string
		ReduceOnly bool
		Ioc bool
		PostOnly bool
		ClientId string
	}
	type res struct {
		Success bool
		Result result
	}
	url := host + orderAPI
	orderStr := order.GetJSONString()
	header := ftx.genAuthHeader("POST", orderAPI, orderStr)
	var resObj res
	ftx.restClient.Post(url, header, order.GetBuffer(), &resObj)
	if !resObj.Success {
		fmt.Println(resObj)
		util.Error(ftx.tag, "Make Order Error")
	}
	return resObj.Result.Id
}