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
	"time"
	character "github.com/CheshireCatNick/crypto-flash/pkg/character"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	"sync"
)

const tag = "Crypto Flash"
// mode: trade, notify, backtest
const mode = "trade"

type user struct {
	Name string
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
	Users []user
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
	
	var n *character.Notifier
	if config.Notify && mode != "backtest" {
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
	ftx := exchange.NewFTX("", "", "")
	sp := character.NewResTrend(ftx, n)
	if mode == "trade" {
		for _, user := range config.Users {
			ftx := exchange.NewFTX(user.Key, user.Secret, user.SubAccount)
			trader := character.NewTrader(user.Name, ftx, n)
			signalChan := make(chan *util.Signal)
			sp.SubSignal(signalChan)
			wg.Add(1)
			go trader.Start(signalChan)
		}
		wg.Add(1)
		go sp.Start()
	} else if mode == "notify" {
		wg.Add(1)
		go sp.Start()
	} else if mode == "backtest" {
		//endTime, _ := time.Parse(time.RFC3339, "2020-03-26T05:00:00+00:00")
		endTime := time.Now()
		d := util.Duration{ Day: -10 }
		startTime := endTime.Add(d.GetTimeDuration())
		roi := sp.Backtest(startTime.Unix(), endTime.Unix())
		annual := util.CalcAnnualFromROI(roi, -d.GetTimeDuration().Seconds())
		fmt.Printf("Annual: %.2f%%", annual * 100)
	}
	wg.Wait()
}
