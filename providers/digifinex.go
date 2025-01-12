package providers

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type FinexMessage struct {
	Event         string   `json:"event"`
	Id            int      `json:"id"`
	Level         int      `json:"level"`
	InstrumentIds string `json:"instrument_id"`
}

type FinexProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewFinexPovider(symbols []string) *FinexProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &FinexProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (c *FinexProvider) GetOrderbooks() orderbook.Orderbooks {
	return c.Orderbooks
}

func (c *FinexProvider) Name() string {
	return "Finex"
}

func (f *FinexProvider) Start() error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://openapi.digifinex.com/swap_ws/v2/", nil)
	if err != nil {
		log.Fatal("DigiFinex websocket dial err:", err)
	}

	//GET ALL COINS
	// msg := FinexMessage{
	// 	Event:         "all_ticker.subscribe",
	// 	Id:            1,
	// }
	for _, symbol := range f.symbols{
		if len(symbol) != 0{
			ws.WriteJSON(FinexMessage{
				Event:         "depth.subscribe",
				Id:            1,
				Level:         10,
				InstrumentIds: symbol,
			})
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second * 50)
		for {
			ws.WriteJSON(FinexMessage{
				Id:    1,
				Event: "server.ping",
			})
			<-ticker.C
		}
	}()

	go func() {
		for {
			_, b, err := ws.ReadMessage()
			if err != nil {
				log.Fatal("DigiFinex read message err:", err)
			}

			buf := bytes.NewReader(b)
			r, err := zlib.NewReader(buf)
			if err != nil {
				log.Fatal("DigiFinex zlib err:", err)
			}
			//DEBUG
			// bb, _ := io.ReadAll(r)
			// log.Println(string(bb))
			//END DEBUG

			msg := FinexSocketResponse{}
			if err := json.NewDecoder(r).Decode(&msg); err != nil {
				continue
				// log.Printf("Error response Unmarshal DigiFinex: %v", err)
				// continue
			}
			
			//DEBUG
			// log.Println(msg.Data)

			if msg.Event == "depth.update" {
				book := f.Orderbooks[msg.Data.InstrumentId]
				for _, ask := range msg.Data.Asks {
					price, size := parseFinexSnapShotEntry(ask)
					book.Asks.Update(price, size)
				}
				for _, bid := range msg.Data.Bids {
					price, size := parseFinexSnapShotEntry(bid)
					book.Bids.Update(price, size)
				}
			}

			// log.Println(msg.Arg.InstId, msg.Data[0].Asks)
		}
	}()

	return nil
}

func parseFinexSnapShotEntry(entry FinexEntry) (float64, float64) {
	p := entry[0].(string)
	price, _ := strconv.ParseFloat(p, 64)
	size := entry[1].(float64)
	return price, size
}

type FinexSocketResponse struct {
	Event string        `json:"event"`
	Data  *ResponseData `json:"data"`
}

type ResponseData struct {
	InstrumentId string       `json:"instrument_id"`
	Asks         []FinexEntry `json:"asks"`
	Bids         []FinexEntry `json:"bids"`
}

type FinexEntry []interface{}

// type FinexEntry struct {
// 	o string
// 	i int
// }
