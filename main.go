/*
// The main program of crypto flash.
*/
package main

import (
	//"encoding/json"
	//"fmt"
	//"io/ioutil"
	//"log"
	//"math"
	//"net/http"
	//exchange "github.com/CheshireCatNick/crypto-flash/pkg/exchange"
	//util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	//"time"
	character "github.com/CheshireCatNick/crypto-flash/pkg/character"
	//cryptoflash "github.com/CheshireCatNick/crypto-flash/pkg/crypto-flash"
)
/*
type config struct {
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
		log.Fatal(err)
	}
	json.Unmarshal(bytes, &c)
	return c
}
*/
func main() {
	//config := loadConfig("config.json")
	sp := character.NewSignalProvider("BTC-PERP", 300)
	/*
	endTime := time.Now()
	var d util.Duration
	d.Day = -1
	startTime := endTime.Add(d.GetTimeDuration())
	sp.Backtest(startTime.Unix(), endTime.Unix())
	*/
	
	sp.Start()

	// test line bot function
	//notifier := cryptoflash.NewNotifier(config.Channel_Secret, config.Channel_Access_Token)
	//notifier.Broadcast("test")
}
