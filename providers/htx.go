package providers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type HTXMessage struct {
	Id  string `json:"id"`
	Sub string `json:"sub"`
}

type HTXProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewHTXProvider(symbols []string) *HTXProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &HTXProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (h *HTXProvider) GetOrderbooks() orderbook.Orderbooks {
	return h.Orderbooks
}

func (h *HTXProvider) Name() string {
	return "HTX"
}

func (h *HTXProvider) Start() error {
	//TODO for all providers
	// u := url.URL{Scheme: "wss", Host: "api.huobi.pro", Path: "/ws"}
	// ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	ws, _, err := websocket.DefaultDialer.Dial("wss://api.huobi.pro/feed", nil)
	if err != nil {
		log.Println("htx websocket dial err:", err)
	}

	for _, symbol := range h.symbols {
		if len(symbol) != 0 {
			ws.WriteJSON(HTXMessage{
				Id:  "id1",
				Sub: symbol,
			})
		}
	}

	ws.ReadMessage()

	go func() {
		for {
			_, rawByte, _ := ws.ReadMessage()

			buf := bytes.NewReader(rawByte)
			r, err := gzip.NewReader(buf)
			if err != nil {
				log.Fatal("htx gzip unpack err:", err)
				continue
			}

			msg := HTXSocketResponse{}
			if err := json.NewDecoder(r).Decode(&msg); err != nil {
				log.Printf("htx unmarshal: %v", err)
				continue
			}

			if msg.Ping != 0 {
				// log.Printf("%+v", msg)
				ws.WriteJSON(
					struct {
						Pong int `json:"pong"`
					}{Pong: msg.Ping})
				continue
			}

			if msg.Tick != nil {

				// log.Print(msg)

				book := h.Orderbooks[msg.Ch]
				for _, ask := range msg.Tick.Asks {
					price, size := parseHTXSnapShotEntry(ask)
					book.Asks.Update(price, size)
				}
				for _, bid := range msg.Tick.Bids {
					price, size := parseHTXSnapShotEntry(bid)
					book.Bids.Update(price, size)
				}

			}

			//DEBUG
			// log.Println(msg.Arg.InstId, msg.Data[0].Asks)
		}
	}()

	return nil
}

func parseHTXSnapShotEntry(entry EntryHTX) (float64, float64) {
	return entry[0], entry[1]
}

type HTXSocketResponse struct {
	Ping int              `json:"ping,omitempty"`
	Ch   string           `json:"ch,omitempty"`
	Tick *HTXResponseTick `json:"tick,omitempty"`
}

type HTXResponseTick struct {
	SeqNum     int        `json:"seqNum"`
	PrevSeqNum int        `json:"prevSeqNum"`
	Asks       []EntryHTX `json:"asks,omitempty"`
	Bids       []EntryHTX `json:"bids,omitempty"`
}

type EntryHTX [2]float64

type HTXMessageUnsub struct {
	Id    string `json:"id"`
	Unsub string `json:"unsub"`
}

func (h *HTXProvider) UnsubscribeHTX(ws *websocket.Conn) {
	for _, symbol := range h.symbols {
		if len(symbol) != 0 {
			ws.WriteJSON(HTXMessageUnsub{
				Id:    "id1",
				Unsub: symbol,
			})
		}
	}
	log.Print("htx unsubscribe done")
}
