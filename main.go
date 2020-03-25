/*
// TODO:
// 1. tests
// 2. consider having exchange interface, signal provider interface
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	//"time"
	character "github.com/CheshireCatNick/crypto-flash/pkg/character"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	"sync"
)

const tag = "Crypto Flash"
type ftxConfig struct {
	Key string
	Secret string
	SubAccount string
}
type lineConfig struct {
	Channel_Secret string
	Channel_Access_Token string
}
type config struct {
	Notify bool
	Ftx ftxConfig
	Line lineConfig
	Telegram string
	Version float32
	Update string
}

func loadConfig(fileName string) config {
	var c config
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		util.Error(tag, err.Error())
	}
	json.Unmarshal(bytes, &c)
	return c
}

func main() {
	var wg sync.WaitGroup
	config := loadConfig("config.json")
	fmt.Printf("Crypto Flash v%.1f initialized. Update: \n%s\n", 
		config.Version, config.Update)
	
	ftx := exchange.NewFTX(config.Ftx.Key, config.Ftx.Secret, 
		config.Ftx.SubAccount)
	
	var n *character.Notifier
	if (config.Notify) {
		n = character.NewNotifier(config.Line.Channel_Secret, 
			config.Line.Channel_Secret, config.Telegram)
		wg.Add(1)
		go n.Listen()
		n.Broadcast(tag, 
			fmt.Sprintf("Crypto Flash v%.1f initialized. Update: \n%s", 
			config.Version, config.Update))
	} else {
		n = nil
	}

	sp := character.NewResTrend(ftx, n)
	trader := character.NewTrader(ftx, n)
	signalChan := make(chan *util.Signal)

	wg.Add(1)
	go sp.Start(signalChan)
	wg.Add(1)
	go trader.Start(signalChan)
	/*
	endTime := time.Now()
	d := util.Duration{ Day: -15 }
	startTime := endTime.Add(d.GetTimeDuration())
	roi := sp.Backtest(startTime.Unix(), endTime.Unix())
	annual := util.CalcAnnualFromROI(roi, -d.GetTimeDuration().Seconds())
	fmt.Printf("Annual: %.2f%%", annual * 100)
	*/
	wg.Wait()
}
