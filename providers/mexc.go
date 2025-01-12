package providers

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type MEXCMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type MEXCProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewMEXCProvider(symbols []string) *MEXCProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &MEXCProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (c *MEXCProvider) GetOrderbooks() orderbook.Orderbooks {
	return c.Orderbooks
}

func (c *MEXCProvider) Name() string {
	return "MEXC"
}

func (m *MEXCProvider) Start() error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://wbs.mexc.com/ws", nil)
	if err != nil {
		log.Fatal("MEXC websocket dial err:", err)
	}

	msg := MEXCMessage{
		Method: "SUBSCRIPTION",
		Params: m.symbols,
	}

	if err = ws.WriteJSON(msg); err != nil {
		log.Fatal("MEXC WriteJSON err:", err)
	}
	ws.ReadMessage()

	go func() {
		for {
			msg := MEXCSocketResponse{}
			if err := ws.ReadJSON(&msg); err != nil {
				log.Fatal("MEXC readJSON err:", err)
				break
			}

			book := m.Orderbooks[msg.C]
			if msg.D.Asks != nil {
				for _, ask := range msg.D.Asks {
					price, size := parseMEXCSnapShotEntry(ask.P, ask.V)
					book.Asks.Update(price, size)
				}
			} else {
				for _, bid := range msg.D.Bids {
					price, size := parseMEXCSnapShotEntry(bid.P, bid.V)
					book.Bids.Update(price, size)
				}
			}

			// if msg.D.Bids != nil {
			// 	for _, bid := range msg.D.Bids {
			// 		price, size := parseMEXCSnapShotEntry(bid.P, bid.V)
			// 		book.Asks.Update(price, size)
			// 		// log.Println(price, size)
			// 	}
			// }
		}

		// log.Println(m.Orderbooks)
	}()

	return nil
}

func parseMEXCSnapShotEntry(p, v string) (float64, float64) {
	price, _ := strconv.ParseFloat(p, 64)
	size, _ := strconv.ParseFloat(v, 64)
	return price, size
}

type MEXCSocketResponse struct {
	C string         `json:"c"`
	D *ResponseDeals `json:"d"`
	S string         `json:"s"`
}

type ResponseDeals struct {
	Asks []*DealsData `json:"asks,omitempty"`
	Bids []*DealsData `json:"bids,omitempty"`
	E    string       `json:"e"`
}

type DealsData struct {
	P string `json:"p"`
	V string `json:"v"`
}
