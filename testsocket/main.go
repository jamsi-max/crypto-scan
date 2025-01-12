package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/socket"
)

func main()  {
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:4000", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	msg := socket.Message{
		Type: "subscribe",
		Topic: "spreads",
		Symbols: []string{"BTCUSD"},
	}

	if err := ws.WriteJSON(msg); err != nil {
		log.Fatal(err)
	}

	for {
		msgg := socket.MessageSpreads{}
		if  err := ws.ReadJSON(&msgg); err != nil {
			log.Fatal(err)
		}
		if len(msgg.Spreads) > 0 {
			fmt.Printf("ex %s ask %f %f bid ex %s sp: %f\r", msgg.Spreads[0].BestAsk.Provider, msgg.Spreads[0].BestAsk.Price, msgg.Spreads[0].BestBid.Price, msgg.Spreads[0].BestBid.Provider, msgg.Spreads[0].Spread)
		}
		// fmt.Println(msg)
	}
}