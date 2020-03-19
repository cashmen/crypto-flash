/*
// The main program of crypto flash.
*/
package main

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
	//"time"
	character "github.com/CheshireCatNick/crypto-flash/pkg/character"
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
	var sp *character.SignalProvider
	if (config.Notify) {
		nf := character.NewNotifier(config.Channel_Secret, 
			config.Channel_Access_Token)
		nf.Broadcast("Crypto Flash initialized.")
		sp = character.NewSignalProvider(nf)
	} else {
		sp = character.NewSignalProvider(nil)
	}
	sp.Start()
	
	/*
	sp := character.NewSignalProvider(nil)
	endTime := time.Now()
	d := util.Duration{ Day: -3 }
	startTime := endTime.Add(d.GetTimeDuration())
	sp.Backtest(startTime.Unix(), endTime.Unix())
	*/
}
