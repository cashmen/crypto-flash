package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	cryptoflash "github.com/CheshireCatNick/crypto-flash/pkg/crypto-flash"
)

type config struct {
	Key        string
	Secret     string
	SubAccount string
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
}
