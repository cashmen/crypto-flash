/*
// Notifier sends messages to users. It can be used to broadcast trade signals
// or send trade operation and ROI notifications.
// Input: signal provider or trader
// Output: 
// TODO: 
// 1. save room ID to DB
*/
package character

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	util "github.com/CheshireCatNick/crypto-flash/pkg/util"
)

type Notifier struct {
	tag string
	lineClient *linebot.Client
	tgClient *tg.BotAPI
	users map[string]int64
}

func NewNotifier(secret, accessToken, tgToken string) *Notifier {
	lc, err := linebot.New(secret, accessToken)
	if err != nil {
		util.Error("Notifier", err.Error())
	}
	tgc, err := tg.NewBotAPI(tgToken)
	if err != nil {
		util.Error("Notifier", err.Error())
	}
	util.Success("Notifier", "auth succeeded", tgc.Self.UserName)
	n := &Notifier{
		tag: "Notifier",
		lineClient: lc,
		tgClient: tgc,
		users: make(map[string]int64),
	}
	n.users["kuroiro_sagishi"] = 441247007
	n.users["liverpool1026"] = 1023854566
	n.users["twoblade"] = 928075336
	return n
}

func (n *Notifier) lineBroadcast(message string) {
	var messages []linebot.SendingMessage
	messages = append(messages, linebot.NewTextMessage(message))
	_, err := n.lineClient.BroadcastMessage(messages...).Do()
	if err != nil {
		util.Error(n.tag, err.Error())
	}
	// TODO: move this to DB
	roomID := []string{
		// bulbul
		//"R129f4d8f3dd39d852d6604b7332c47fa",
	}
	for _, rID := range roomID {
		_, err := n.lineClient.PushMessage(rID, messages...).Do()
		if err != nil {
			util.Error(n.tag, err.Error())
		}
	}
}
func (n *Notifier) Listen() {
	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates, err := n.tgClient.GetUpdatesChan(u)
	if err != nil {
		util.Error(n.tag, err.Error())
	}
	// receive from tg bot
	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}
			recvMsg := update.Message
			msg := tg.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "Available commands: /start, /register and /status."
			case "start":
				user := recvMsg.From.UserName
				chatID := recvMsg.Chat.ID
				n.users[user] = chatID
				util.Success(n.tag, "register", user, util.PI64(chatID))
				msg.Text = "Welcome to Crypto Flash, " + user + 
					". Please note that signals from this bot are " + 
					"recommendations and you should trade on your own " + 
					"responsibility. Enjoy!"
			case "register":
				user := recvMsg.From.UserName
				chatID := recvMsg.Chat.ID
				msg.Text = "Hi, " + user + " : )"
				n.users[user] = chatID
				util.Success(n.tag, "register", user, util.PI64(chatID))
			case "status":
				msg.Text = "I'm ok."
			default:
				msg.Text = "I don't know the command."
			}
			n.tgClient.Send(msg)
		}
	}()
}
func (n *Notifier) tgBroadcast(message string) {
	for _, chatID := range n.users {
		msg := tg.NewMessage(chatID, message)
		n.tgClient.Send(msg)
	}
}
func (n *Notifier) tgSend(to, message string) {
	msg := tg.NewMessage(n.users[to], message)
	n.tgClient.Send(msg)
}
func (n *Notifier) Broadcast(from, message string) {
	n.tgBroadcast(fmt.Sprintf("[%s] %s", from, message))
}
func (n *Notifier) Send(from, to, message string) {
	n.tgSend(to, fmt.Sprintf("[%s] %s", from, message))
}
