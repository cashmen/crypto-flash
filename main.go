/*
// The main program of crypto flash.
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

const tag = "Main"
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
	
	ftx := exchange.NewFTX(config.Ftx.Key, config.Ftx.Secret, 
		config.Ftx.SubAccount)
	wallet := ftx.GetWallet()
	fmt.Println(wallet)
	/*
	order := &util.Order{
		Market: "BTC-PERP",
		Side: "buy",
		Price: 1000,
		Type: "limit",
		Size: 0.0001,
		ClientId: nil,
	}*/
	//ftx.MakeOrder(order)
	wg.Add(1)
	
	var sp *character.SignalProvider
	if (config.Notify) {
		n := character.NewNotifier(config.Line.Channel_Secret, 
			config.Line.Channel_Secret, config.Telegram)
		n.Broadcast("Crypto Flash initialized.")
		sp = character.NewSignalProvider(ftx, n)
	} else {
		sp = character.NewSignalProvider(ftx, nil)
	}
	sp.Start()
	
	/*
	sp := character.NewSignalProvider(ftx, nil)
	endTime := time.Now()
	d := util.Duration{ Day: -3 }
	startTime := endTime.Add(d.GetTimeDuration())
	sp.Backtest(startTime.Unix(), endTime.Unix())
	*/
	wg.Wait()
}
