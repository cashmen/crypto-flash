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
	//character "github.com/CheshireCatNick/crypto-flash/pkg/character"
	exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
)

const tag = "Main"
type config struct {
	Notify bool
	Key        string
	Secret     string
	SubAccount string
	Channel_Secret string
	Channel_Access_Token string
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
	
	config := loadConfig("config.json")
	ftx := exchange.NewFTX(config.Key, config.Secret, config.SubAccount)
	wallet := ftx.GetWallet()
	fmt.Println(wallet)
	order := &util.Order{
		Market: "BTC-PERP",
		Side: "buy",
		Price: 1000,
		Type: "limit",
		Size: 0.0001,
		ClientId: nil,
	}
	ftx.MakeOrder(order)

	/*
	var sp *character.SignalProvider
	if (config.Notify) {
		nf := character.NewNotifier(config.Channel_Secret, 
			config.Channel_Access_Token)
		nf.Broadcast("Crypto Flash initialized.")
		sp = character.NewSignalProvider(ftx, nf)
	} else {
		sp = character.NewSignalProvider(ftx, nil)
	}
	sp.Start()*/
	
	/*
	sp := character.NewSignalProvider(ftx, nil)
	endTime := time.Now()
	d := util.Duration{ Day: -3 }
	startTime := endTime.Add(d.GetTimeDuration())
	sp.Backtest(startTime.Unix(), endTime.Unix())
	*/
}
