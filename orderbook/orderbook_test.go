package orderbook

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestOrderbookBestAskBid(t *testing.T) {
// 	var (
// 		ob = NewOrderbook("foobar")
// 		n = 1000
// 	)

// 	for i := 1; i< n; i++ {
// 		aorder := NewAskOrder(float64(i))
// 		ob.placeLimitOrder(float64(i), aorder)

// 		border := NewBidOrder(float64(i))
// 		ob.placeLimitOrder(float64(i), border)
// 	}

// 	assert.Equal(t, 1.0, ob.bestAsk().price)
// 	assert.Equal(t, 999.0, ob.bestBid().price)
// }

// func TestOrderBookPlaceLimitOrder(t *testing.T) {
// 	var (
// 		ob = NewOrderbook("foobar")
// 		n = 1000
// 	)

// 	size := 0.0
// 	for i:= 1; i < n; i++ {
// 		order := NewAskOrder(float64(i))
// 		require.Equal(t, false, order.isBid)
// 		ob.placeLimitOrder(float64(i), order)
// 		size += order.size
// 		assert.Equal(t, i, ob.asks.data.Len())
// 		assert.Equal(t, size, ob.totalAskVolume())
// 		assert.Equal(t, 1.0, ob.bestAsk(),price)
// 	}

// 	size = 0.0
// 	for i:= 1; i < n; i++ {
// 		order := NewBidOrder(float64(i))
// 		require.Equal(t, true, order.isBid)
// 		ob.placeLimitOrder(float64(i), order)
// 		size += order.size
// 		assert.Equal(t, i, ob.bids.data.Len())
// 		assert.Equal(t, size, ob.totalBidVolume())
// 		assert.Equal(t, 1.0, ob.bestBid(),price)
// 	}
// }

// func TestNewOrderbook(t *testing.T) {
// 	asksFile := "../data/asks.gob"
// 	bidsFile := "../data/bids.gob"
// 	ob, err := NewOrderbookFromFile("BTCUSDT", asksFile, bidsFile)
// 	assert.Nil(t, err)

// 	for i := 0; i < 10; i++ {
// 		fmt.Println("asks volume:", ob.totalAskVolume())
// 		fmt.Println("asks volume:", ob.totalBidVolume())
// 	}
// }

// func TestLimitFillMultiOrder(t *testing.T) {
// 	l := NewLimit(10_000)
// 	askOrderA := NewAskOrder(10)
// 	askOrderB := NewAskOrder(5)
// 	l.addOrder(askOrderA)
// 	l.addOrder(askOrderB)

// 	marketOrder := NewAskOrder(12)
// 	l.fillOrder(marketOrder)
// 	assert.Equal(t, 3.0, l.totalVolume)

// 	assert.Equal(t, 1, len(l.orders))
// 	assert.Equal(t, 3.0, l.orders[0].size)
// 	assert.True(t, marketOrder.isFilled())
// }

// func TestLimitFillSingleOrder(t *testing.T) {
// 	var (
// 		l 						  = NewLimit(50_000)
// 		orderSize       = 10.0
// 		marketOrderSize = 6.0
// 		askOrder        = NewAskOrder(orderSize)
// 		marketOrder     = NewAskOrder(marketOrderSize)
// 	)

// 	l.addOrder(askOrder)
// 	l.addOrder(askOrder)
// 	l.fillOrder(marketOrder)
// 	assert.True(t, marketOrder.isFilled())
// 	assert.Equal(t, orderSize-marketOrderSize, askOrder.size)
// 	assert.Equal(t, askOrder.size, l.totalVolume)
// }

// func TestLimitDeleteOrder(t *testing.T) {
// 	l := NewLimit(20_000)
// 	o1 := NewBidOrder(1.0)
// 	o2 := NewBidOrder(2.0)
// 	o3 := NewBidOrder(3.0)
// 	l.addOrder(o1)
// 	l.addOrder(o2)
// 	l.addOrder(o3)
// 	assert.Equal(t, 6.0, l.totalVolume)

// 	l.deleteOrder(o2)
// 	assert.Equal(t, 4.0, l.totalVolume)
// }

// func TestLimit_AddOrder(t *testing.T)  {
// 	l := NewLimit(16000)
// 	n := 1000
// 	size := 50.0

// 	for i:= 0; i < n; i++ {
// 		o := NewAskOrder(size)
// 		l.addOrder(o)
// 		assert.Equal(t, n, o.limitIndex)
// 	}

// 	assert.Equal(t, n, len(l.orders))
// 	assert.Equal(t, float64(n)*size, l.totalVolume)
// }