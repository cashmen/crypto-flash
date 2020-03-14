package cobinhood

import (
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"time"
)

type WSClient struct {
	isConnected	bool
	terminal	bool
	connection	websocket.Dialer
}

func (c *WSClient) subscribeBooks() {

}

func New() *WSClient {
	c := &WSClient{
		isConnected:	false,
		terminal:		false,
		connection:		websocket.Dialer{},
	}
	return c
}

func main() {

	//
	client := New()
	client.subscribeBooks()
	//

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	addr := "ws.cobinhood.com"
	u := url.URL{Scheme: "wss", Host: addr, Path: "/v2/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			log.Printf("tick")
			err := c.WriteMessage(websocket.TextMessage, []byte("{\"action\": \"ping\"}"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
