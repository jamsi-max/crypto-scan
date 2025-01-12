package providers

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/jamsi-max/arbitrage/orderbook"
)

type CoinbaseProvider struct {
	Orderbooks orderbook.Orderbooks
	symbols    []string
}

func NewCoinbaseProvider(symbols []string) *CoinbaseProvider {
	books := orderbook.Orderbooks{}
	for _, symbol := range symbols {
		books[symbol] = orderbook.NewBook(symbol)
	}

	return &CoinbaseProvider{
		Orderbooks: books,
		symbols:    symbols,
	}
}

func (c *CoinbaseProvider) GetOrderbooks() orderbook.Orderbooks {
	return c.Orderbooks
}

func (c *CoinbaseProvider) Name() string {
	return "Coinbase"
}

func (c *CoinbaseProvider) Start() error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	msg := CoinbaseMessage{
		Type:       "subscribe",
		ProductIds: c.symbols,
		Channels:   []string{"level2_batch"}, // level2 требует auth через сайт и KYC для получения ключа
	}
	if err = ws.WriteJSON(msg); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			msg := CoinbaseSocketResponse{}
			if err := ws.ReadJSON(&msg); err != nil {
				log.Fatal(err)
				break
			}

			if msg.Type == "errorsubscriptions" {
				continue
			}

			if msg.Type == "l2update" {
				c.handleUpdate(msg.ProductID, msg.Changes)
			}
			if msg.Type == "snapshot" {
				c.handleSnapshot(msg.ProductID, msg.Asks, msg.Bids)
			}
		}
	}()

	return nil
}

func (c *CoinbaseProvider) handleUpdate(symbol string, changes []SnapshotChange) error {
	for _, change := range changes {
		side, price, size := parseSnapShotChange(change)
		if side == "sell" {
			c.Orderbooks[symbol].Asks.Update(price, size)
		} else {
			c.Orderbooks[symbol].Bids.Update(price, size)
		}

	}

	return nil
}

func (c *CoinbaseProvider) handleSnapshot(symbol string, asks []SnapshotEntry, bids []SnapshotEntry) error {
	for _, entry := range asks {
		price, size := parseSnapShotEntry(entry)
		c.Orderbooks[symbol].Asks.Update(price, size)
	}
	for _, entry := range bids {
		price, size := parseSnapShotEntry(entry)
		c.Orderbooks[symbol].Bids.Update(price, size)
	}
	return nil
}

func parseSnapShotChange(change SnapshotChange) (string, float64, float64) {
	side := change[0]
	price, _ := strconv.ParseFloat(change[1], 64)
	size, _ := strconv.ParseFloat(change[2], 64)
	return side, price, size
}

func parseSnapShotEntry(entry [2]string) (float64, float64) {
	price, _ := strconv.ParseFloat(entry[0], 64)
	size, _ := strconv.ParseFloat(entry[1], 64)
	return price, size
}

type CoinbaseMessage struct {
	Type       string   `json:"type"`
	ProductIds []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type CoinbaseSocketResponse struct {
	Type       string   `json:"type"`
	ProductID  string   `json:"product_id"`
	ProductIds []string `json:"product_ids"`

	TradeID      int    `json:"trade_id"`
	OrderID      string `json:"order_id"`
	ClientOID    string `json:"client_oid"`
	Sequence     int64  `json:"sequence"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`

	RemainigSize string `json:"remainig_size"`
	NewSize      string `json:"new_size"`
	OldSize      string `json:"old_size"`

	Reason      string           `json:"reason"`
	OrderType   string           `json:"order_type"`
	Funds       string           `json:"funds"`
	NewFunds    string           `json:"new_funds"`
	OldFunds    string           `json:"old_funds"`
	Message     string           `json:"message"`
	Bids        []SnapshotEntry  `json:"bids,omitempty"`
	Asks        []SnapshotEntry  `json:"asks,omitempty"`
	Changes     []SnapshotChange `json:"changes,omitempty"`
	LastSize    string           `json:"last_size"`
	BestBid     string           `json:"best_bid"`
	BestAsk     string           `json:"best_ask"`
	Channels    []MessageChannel `json:"channels"`
	UserID      string           `json:"user_id"`
	ProfileID   string           `json:"profile_id"`
	LastTradeID int              `json:"last_trade_id"`
}

type MessageChannel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

type SnapshotChange [3]string

type SnapshotEntry [2]string
