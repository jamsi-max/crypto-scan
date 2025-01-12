package providers

import (
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type GateioProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewGateioProvider(symbols []string) *GateioProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &GateioProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (c *GateioProvider) GetOrderbooks() orderbook.Orderbooks {
	return c.Orderbooks
}

func (c *GateioProvider) Name() string {
	return "Gateio"
}

func (g *GateioProvider) Start() error {
	//TODO for all providers
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("gateio websocket Dial err:", err)
	}
	
	ws.SetPingHandler(nil)

	for _, symbol := range g.symbols {
		err := ws.WriteJSON(GateioMessageSubscribe{
			Time:    time.Now().Unix(),
			Channel: "spot.order_book_update",
			Event:   "subscribe",
			Payload: []string{symbol, "100ms"},
		})
		if err != nil {
			log.Printf("gateio subscribe %v err: %v", symbol, err)
		}
	}

	go func() {
		for {
			msg := GateioSocketResponse{}
			if err := ws.ReadJSON(&msg); err != nil {
				log.Fatal("gateio readJSON err:", err)
				break
			}
			
			if msg.Event == "update" {
				book := g.Orderbooks[msg.Result.S]
				for _, ask := range msg.Result.A {
					price, size := parseGateioSnapShotEntry(ask)
					book.Asks.Update(price, size)
				}
				for _, bid := range msg.Result.B {
					price, size := parseGateioSnapShotEntry(bid)
					book.Bids.Update(price, size)
				}
			}
			//DEBUG
			// log.Println(msg.Arg.InstId, msg.Data[0].Asks)
		}
	}()

	return nil
}

func parseGateioSnapShotEntry(entry GateioEntry) (float64, float64) {
	price, _ := strconv.ParseFloat(entry[0], 64)
	size, _ := strconv.ParseFloat(entry[1], 64)
	return price, size
}

type GateioMessageSubscribe struct {
	Time    int64    `json:"time"`
	Channel string   `json:"channel"`
	Event   string   `json:"event"`
	Payload []string `json:"payload"`
}

type GateioSocketResponse struct {
	Event  string           `json:"event"`
	Result *GateioOrderbook `json:"result"`
}

type GateioOrderbook struct {
	S    string        `json:"s"`
	B []GateioEntry `json:"b"`
	A []GateioEntry `json:"a"`
}

type GateioEntry [2]string
