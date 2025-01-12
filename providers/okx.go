package providers

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type OKXProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewOKXProvider(symbols []string) *OKXProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &OKXProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (c *OKXProvider) GetOrderbooks() orderbook.Orderbooks {
	return c.Orderbooks
}

func (c *OKXProvider) Name() string {
	return "OKX"
}

func (o *OKXProvider) Start() error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://ws.okx.com:8443/ws/v5/public", nil)
	if err != nil {
		log.Fatal(err)
	}

	args := make([]OKXMessageSubscribeArg, len(o.symbols))
	for i, symbol := range o.symbols {
		args[i] = OKXMessageSubscribeArg{
			Channel: "books",
			InstId:  symbol,
		}
	}
	msg := OKXMessageSubscribe{
		Op:   "subscribe",
		Args: args,
	}

	if err = ws.WriteJSON(msg); err != nil {
		log.Fatal(err)
	}
	ws.ReadMessage()

	go func() {
		for {
			msg := OKXSocketResponse{}
			if err := ws.ReadJSON(&msg); err != nil {
				log.Fatal("OKX readJSON err:", err)
				break
			}

			if msg.Action == "update" {
				book := o.Orderbooks[msg.Arg.InstId]
				for _, dataElms := range msg.Data {
					for _, asks := range dataElms.Asks {
						price, size := parseOKXSnapShotEntry(asks)
						book.Asks.Update(price, size)
					}
					for _, bids := range dataElms.Bids {
						price, size := parseOKXSnapShotEntry(bids)
						book.Bids.Update(price, size)
					}
				}
				
			}
			// log.Println(msg.Arg.InstId, msg.Data[0].Asks)
		}
	}()
	
	return nil
}

func parseOKXSnapShotEntry(entry OKXEntry) (float64, float64) {
	price, _ := strconv.ParseFloat(entry[0], 64)
	size, _ := strconv.ParseFloat(entry[1], 64)
	return price, size
}

type OKXMessageSubscribe struct {
	Op   string                   `json:"op"`
	Args []OKXMessageSubscribeArg `json:"args"`
}

type OKXMessageSubscribeArg struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type OKXSocketResponse struct {
	Arg    *ResponseArg   `json:"arg"`
	Action string         `json:"action"`
	Data   []OKXOrderbook `json:"data"`
}

type ResponseArg struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type OKXOrderbook struct {
	Asks []OKXEntry `json:"asks"`
	Bids []OKXEntry `json:"bids"`
}

type OKXEntry [4]string
