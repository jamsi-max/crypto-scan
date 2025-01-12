package orderbook

import (
	"encoding/gob"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/VictorLowther/btree"
)

type BestPrice struct {
	Provider string
	Price    float64
	Size     float64
}

type CrossSpread struct {
	Symbol  string
	BestAsk BestPrice
	BestBid BestPrice
	Spread  float64
}

type Provider interface {
	Start() error
	GetOrderbooks() Orderbooks
	Name() string
}

type Orderbooks map[string]*Book

type Book struct {
	Symbol string
	Asks   *Limits
	Bids   *Limits
}

func NewBook(symbol string) *Book {
	return &Book{
		Symbol: symbol,
		Asks:   NewLimits(false),
		Bids:   NewLimits(true),
	}
}

func (b *Book) Spread() float64 {
	if b.Asks.data.Len() == 0 || b.Bids.data.Len() == 0 {
		return 0.0
	}
	bestAsk := b.Asks.Best().Price
	bestBid := b.Bids.Best().Price
	return bestAsk - bestBid
}

func (b *Book) BestBid() *Limit {
	return b.Bids.Best()
}

func (b *Book) BestAsk() *Limit {
	return b.Asks.Best()
}

func getBidByPrice(price float64) btree.CompareAgainst[*Limit] {
	return func(l *Limit) int {
		switch {
		case l.Price > price:
			return -1
		case l.Price < price:
			return 1
		default:
			return 0
		}
	}
}

func getAskByPrice(price float64) btree.CompareAgainst[*Limit] {

	return func(l *Limit) int {
		switch {
		case l.Price < price:
			return -1
		case l.Price > price:
			return 1
		default:
			return 0
		}
	}
}

func sortByBestBid(a, b *Limit) bool {
	return a.Price > b.Price
}

func sortByBestAsk(a, b *Limit) bool {
	return a.Price < b.Price
}

type Limits struct {
	isBids      bool
	lock        sync.RWMutex
	data        *btree.Tree[*Limit]
	totalVolume float64
}

func NewLimits(isBids bool) *Limits {
	f := sortByBestAsk
	if isBids {
		f = sortByBestBid
	}
	return &Limits{
		isBids: isBids,
		data:   btree.New(f),
	}
}

func (l *Limits) Len() int {
	return l.data.Len()
}

func (l *Limits) Best() *Limit {
	l.lock.RLock()
	defer l.lock.RUnlock()

	if l.data.Len() == 0 {
		return nil
	}
	iter := l.data.Iterator(nil, nil)
	iter.Next()
	return iter.Item()
}

func (l *Limits) Update(price float64, size float64) {
	l.lock.Lock()
	defer l.lock.Unlock()

	getFunc := getAskByPrice(price)
	if l.isBids {
		getFunc = getBidByPrice(price)
	}

	if limit, ok := l.data.Get(getFunc); ok {
		//
		if size == 0.0 {
			//DEBUG
			// fmt.Println(price, limit.TotalVolume)

			l.data.Delete(limit)
			return
		}
		//DEBUG
		// fmt.Println(price, limit.TotalVolume)
		limit.TotalVolume = size
		return
	}

	if size == 0.0 {
		return
	}

	limit := NewLimit(price)
	limit.TotalVolume = size
	l.data.Insert(limit)
}

type BestSpread struct {
	Symbol  string
	A       string
	B       string
	BestBid float64
	BestAsk float64
	Spread  float64
}

type DataFeed struct {
	Provider string
	Symbol   string
	BestAsk  float64
	BestBid  float64
	Spread   float64
}

func (l *Limits) addOrder(price float64, o *Order) {
	if o.isBid != l.isBids {
		panic("the side of the limits does not match the side of the odrer")
	}

	f := getAskByPrice(price)
	if l.isBids {
		f = getBidByPrice(price)
	}

	var (
		limit *Limit
		ok    bool
	)

	limit, ok = l.data.Get(f)
	if !ok {
		limit = NewLimit(price)
		l.data.Insert(limit)
	}

	l.totalVolume += o.size

	limit.addOrder(o)
}

func (l *Limits) loadFromFile(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}

	var data map[float64]float64
	if err := gob.NewDecoder(f).Decode(&data); err != nil {
		return err
	}

	for price, size := range data {
		limit := NewLimit(price)
		limit.TotalVolume = size

		l.data.Insert(limit)
		l.totalVolume += size
	}

	return nil
}

type Orderbook struct {
	pair string
	asks *Limits
	bids *Limits
}

func NewOrderbookFromFile(pair, askSrc, bidSrc string) (*Orderbook, error) {
	asks := NewLimits(false)
	if err := asks.loadFromFile(askSrc); err != nil {
		return nil, err
	}
	bids := NewLimits(true)
	if err := bids.loadFromFile(bidSrc); err != nil {
		return nil, err
	}

	return &Orderbook{
		pair: pair,
		asks: asks,
		bids: bids,
	}, nil
}

// type BestPrice struct{
// 	Provider string
// 	Price    float64
// 	Size     float64
// }

// type CrossSpread struct{
// 	Symbol  string
// 	BestAsk BestPrice
// 	BestBid BestPrice
// 	Spread  float64
// }

type ByBestAsk struct{ LimitMap }

func (ba ByBestAsk) Len() int { return len(ba.LimitMap.limits) }

type LimitMap struct {
	isBids      bool
	limits      map[float64]*Limit
	totalVolume float64
}

func NewLimitMap(isBids bool) *LimitMap {
	return &LimitMap{
		isBids: isBids,
		limits: make(map[float64]*Limit),
	}
}

func NewOrderbook(pair string) *Orderbook {
	return &Orderbook{
		pair: pair,
	}
}

func (ob *Orderbook) totalAskVolume() float64 {
	return ob.asks.totalVolume
}

func (ob *Orderbook) totalBidVolume() float64 {
	return ob.bids.totalVolume
}

type Limit struct {
	Price       float64
	Orders      []*Order
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

func (l *Limit) fillOrder(marketOrder *Order) {
	ordersToDelete := []*Order{}
	for _, limitOrder := range l.Orders {
		max, min := maxMinOrder(limitOrder, marketOrder)
		sizeFilled := min.size
		max.size -= sizeFilled
		l.TotalVolume -= sizeFilled
		min.size = 0.0

		if limitOrder.isFilled() {
			ordersToDelete = append(ordersToDelete, limitOrder)
		}

		if marketOrder.isFilled() {
			break
		}
	}

	for _, order := range ordersToDelete {
		l.deleteOrder(order)
	}
}

func (l *Limit) addOrder(o *Order) {
	l.Orders = append(l.Orders, o)
	o.limitIndex = len(l.Orders)
	l.TotalVolume += o.size
}

func (l *Limit) deleteOrder(o *Order) {
	l.Orders[o.limitIndex] = l.Orders[len(l.Orders)-1]
	l.Orders = l.Orders[:len(l.Orders)-1]

	if !o.isFilled() {
		l.TotalVolume -= o.size
	}
}

type Order struct {
	id         int64
	size       float64
	timestamp  int64
	isBid      bool
	limitIndex int
}

func NewOrder(isBid bool, size float64) *Order {
	return &Order{
		id:        rand.Int63n(100000),
		size:      size,
		timestamp: time.Now().UnixNano(),
		isBid:     isBid,
	}
}

func NewBidOrder(size float64) *Order {
	return NewOrder(false, size)
}

func NewAskOrder(size float64) *Order {
	return NewOrder(false, size)
}

func (o *Order) isFilled() bool {
	return o.size == 0
}

func maxMinOrder(a, b *Order) (*Order, *Order) {
	if a.size >= b.size {
		return a, b
	}
	return b, a
}
