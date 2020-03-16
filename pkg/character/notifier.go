/*
// Notifier sends messages to users. It can be used to broadcast trade signals
// or send trade operation and ROI notifications.
// Input: signal provider or trader
// Output: 
// TODO: 
// 1. finish send()
// 2. support Telegram
*/
package character

import (
	"log"
	"github.com/line/line-bot-sdk-go/linebot"
)

type Notifier struct {
	//channelSecret string
	//channelAccessToken string
	lineClient *linebot.Client
}

func NewNotifier(secret string, accessToken string) *Notifier {
	c, err := linebot.New(secret, accessToken)
	if err != nil {
		log.Fatal(err)
	}
	n := Notifier{lineClient: c}
	return &n
}

func (n *Notifier)Broadcast(message string) {
	var messages []linebot.SendingMessage
	messages = append(messages, linebot.NewTextMessage(message))
	_, err := n.lineClient.BroadcastMessage(messages...).Do()
	// also send to rooms
	if err != nil {
		// Do something when some bad happened
		log.Fatal(err)
	}
	// TODO: move this to DB
	roomID := []string{
		// rong
		"R6b0164c3372347626ab56afae04b9a79",
		// bulbul
		"R129f4d8f3dd39d852d6604b7332c47fa",
	}
	for _, rID := range roomID {
		_, err := n.lineClient.PushMessage(rID, messages...).Do()
		if err != nil {
			// Do something when some bad happened
			log.Fatal(err)
		}
	}
}
