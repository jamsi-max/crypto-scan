package providers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"time"

	"github.com/Kucoin/kucoin-go-sdk"
	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type TokenModel struct {
	Data *DataTokenModel
}

type DataTokenModel struct {
	Token           string `json:"token"`
	InstanceServers []InstanceServers
}

type InstanceServers struct {
	Endpoint     string        `json:"endpoint"`
	Encrypt      bool          `json:"encrypt"`
	Protocol     string        `json:"protocol"`
	PingInterval time.Duration `json:"pingInterval"`
}

var Timeout time.Duration

type CucoinProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewCucoinProvider(symbols []string) *CucoinProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &CucoinProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (c *CucoinProvider) GetOrderbooks() orderbook.Orderbooks {
	return c.Orderbooks
}

func (c *CucoinProvider) Name() string {
	return "Cucoin"
}

const ApiGetPublickToken = "https://api.kucoin.com/api/v1/bullet-public"

func GetTokenAndEndpoint() (string, string) {
	data := []byte{}

	r := bytes.NewReader(data)
	resp, err := http.Post(ApiGetPublickToken, "application/json", r)
	if err != nil {
		log.Fatalf("Error get Token kukoin: %v", err)
	}
	defer resp.Body.Close()

	var t TokenModel
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		log.Printf("Error response Unmarshal kukoin: %v", err)
	}

	if t.Data.InstanceServers != nil && t.Data.InstanceServers[0].Endpoint != "" {
		Timeout = t.Data.InstanceServers[0].PingInterval
		return t.Data.Token, t.Data.InstanceServers[0].Endpoint
	}

	return t.Data.Token, "wss://ws-api-spot.kucoin.com/"
}

func (c *CucoinProvider) Start() error {
	token, endpoint := GetTokenAndEndpoint()
	connectId := time.Now().UnixNano()
	wsURL := endpoint + "?token=" + token + "&[connectId=" + kucoin.IntToString(connectId) + "]"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	ws.ReadMessage()

	for _, symbol := range c.symbols {
		ws.WriteJSON(MessageSubscribe{
			Id:             connectId,
			Type:           "subscribe",
			Topic:          symbol,
			PrivateChannel: false,
			Response:       true,
		})
	}

	//DEBUG
	// for {
	// 	_, m, _ := ws.ReadMessage()
	// 	fmt.Printf("%s\r", string(m))
	// }
	// select {}
	// END DEBUG

	go func() {
		ticker := time.NewTicker(time.Millisecond * Timeout)
		for {
			ws.WriteJSON(MessageSubscribe{
				Id:   connectId,
				Type: "ping",
			})
			<-ticker.C
		}
	}()

	go func() {
		//SPEED REQUESTS SET
		// ticker := time.NewTicker(time.Millisecond * 100)
		for {
			msg := CucoinSocketResponse{}
			if err := ws.ReadJSON(&msg); err != nil {
				log.Fatal("Cucoin readJSON err:", err)
				break
			}

			if msg.Type == "message" {
				book := c.Orderbooks[msg.Topic]
				for _, asks := range msg.Data.Changes.Asks {
					price, size := parseCucoinSnapShotEntry(asks)
					book.Asks.Update(price, size)
				}
				for _, bids := range msg.Data.Changes.Bids {
					price, size := parseCucoinSnapShotEntry(bids)
					book.Bids.Update(price, size)
				}
			}
			// <-ticker.C
		}
	}()

	return nil
}

func parseCucoinSnapShotEntry(entry CucoinEntry) (float64, float64) {
	price, _ := strconv.ParseFloat(entry[0], 64)
	size, _ := strconv.ParseFloat(entry[1], 64)
	return price, size
}

type MessageSubscribe struct {
	Id             int64  `json:"id"`
	Type           string `json:"type"`
	Topic          string `json:"topic"`
	PrivateChannel bool   `json:"privateChannel"`
	Response       bool   `json:"response"`
}

type CucoinSocketResponse struct {
	// Id    string `json:"id,omitempty"`
	Topic string `json:"topic"`
	Type  string `json:"type"`
	Data  *CucoinOrderbook
}

type CucoinOrderbook struct {
	Changes       *CucoinOrderbookChanges
	SequenceEnd   int    `json:"sequenceEnd"`
	SequenceStart int    `json:"sequenceStart"`
	Symbol        string `json:"symbol"`
	Time          int    `json:"time"`
}

type CucoinOrderbookChanges struct {
	Asks []CucoinEntry `json:"asks"`
	Bids []CucoinEntry `json:"bids"`
}

type CucoinEntry [3]string
