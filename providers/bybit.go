package providers

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type BybitMessage struct {
	Op    string   `json:"op"`
	Args  []string `json:"args"`
}

type BybitProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewBybitProvider(symbols []string) *BybitProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &BybitProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (p *BybitProvider) Name() string {
	return "Bybit"
}

func (p *BybitProvider) GetOrderbooks() orderbook.Orderbooks {
	return p.Orderbooks
}

func (p *BybitProvider) Start() error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://stream.bybit.com/v5/public/spot", nil)
	if err != nil {
		log.Fatal("bybit dial err: ", err)
	}

	msg := BybitMessage{
		Op: "subscribe",
		// Args: []string{"orderbook.1.BTCUSDT", "orderbook.1.LTCUSDT"},
		Args: p.symbols,
	}

	if err = ws.WriteJSON(msg); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			msg := BybitSocketResponse{}
			if err := ws.ReadJSON(&msg); err != nil {
				log.Println("bybit readJSON err:", err)
				continue
			}
			// log.Println("->", msg)

			if msg.Type == "delta" {
				// log.Printf("%v %+v",msg.Topic, msg.Data)   //>>>debug
				book := p.Orderbooks[msg.Topic]
				for _, ask := range msg.Data.A {
					price, size := parseSnapShotEntry(ask)
					book.Asks.Update(price, size)
				}
				for _, bid := range msg.Data.B {
					price, size := parseSnapShotEntry(bid)
					book.Bids.Update(price, size)
				}
			}
		}
	}()

	return nil
}

type BybitSocketResponse struct {
	Topic string         `json:"topic"`
	Type  string         `json:"type"`
	Data  BybitOrderbook `json:"data"`
}

type BybitOrderbook struct {
	S string  `json:"s"`
	B []Entry `json:"b"`
	A []Entry `json:"a"`
}

type Entry [2]string
