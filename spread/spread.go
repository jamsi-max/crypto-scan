package spread

import (
	"github.com/jamsi-max/arbitrage/orderbook"
	sttg "github.com/jamsi-max/arbitrage/settings"
	"github.com/jamsi-max/arbitrage/util"
)

func getSymbolForProvider(p string, symbol string) string {
	return sttg.Pairs[symbol][p]
}

func CalcCrossSpreads(datach chan map[string][]orderbook.CrossSpread, pvrs []orderbook.Provider) {
	data := map[string][]orderbook.CrossSpread{}

	for _, symbol := range sttg.Symbols {
		crossSpreads := []orderbook.CrossSpread{}
		// for i := 0; i < len(pvrs); i++ {
		for i, j := 0, 1; i < len(pvrs)-1; {
			a := pvrs[i]
			b := pvrs[j]

			// a := pvrs[i]
			// var b orderbook.Provider
			// if len(pvrs)-1 == i {
			// 	b = pvrs[0]
			// } else {
			// 	b = pvrs[i+1]
			// }

			var (
				crossSpread = orderbook.CrossSpread{
					Symbol: symbol,
				}
				bestAsk  = orderbook.BestPrice{}
				bestBid  = orderbook.BestPrice{}
				bookA    = a.GetOrderbooks()[getSymbolForProvider(a.Name(), symbol)]
				bookB    = b.GetOrderbooks()[getSymbolForProvider(b.Name(), symbol)]
				bestBidA = bookA.BestBid()
				bestBidB = bookB.BestBid()
			)

			if bestBidA == nil || bestBidB == nil {
				if j < len(pvrs)-1 {
					j++
				} else {
					i++
					j = i + 1
				}
				continue
			}
			// fmt.Println(a.Name(), bestBidA.Price, b.Name(), bestBidB.Price)

			//DEBUG
			// fmt.Printf("%+v", len(pvrs))
			// fmt.Println(i, j, a.Name(), bestBidA.Price, b.Name(), bestBidB.Price)
			// if b.Name() == "Cucoin"  {
			// 	log.Println(b.Name(), bookB.BestAsk().Price)
			// }
			// DEBUG END

			if bestBidA.Price < bestBidB.Price {
				// log.Println("a<b", a.Name(),bestBidA.Price, b.Name(), bestBidB.Price, "b-a:", bestBidB.Price-bestBidA.Price)
				bestBid.Provider = a.Name()
				bestAsk.Provider = b.Name()
				bestBid.Price = bestBidA.Price
				bestBid.Size = bestBidA.TotalVolume
				if bookB.BestAsk() != nil {
					bestAsk.Price = bookB.BestAsk().Price
					bestAsk.Size = bookB.BestAsk().TotalVolume
				}
				// if symbol == "SOLUSD" && (a.Name() == "Cucoin" || b.Name() == "Cucoin"){
				// 	log.Println("a<b",symbol, a.Name(), bestBid.Price, b.Name(), bestAsk.Price)
				// }
			} else {
				bestBid.Provider = b.Name()
				bestAsk.Provider = a.Name()
				bestBid.Price = bestBidB.Price
				bestBid.Size = bestBidB.TotalVolume
				if bookA.BestAsk() != nil {
					bestAsk.Price = bookA.BestAsk().Price
					bestAsk.Size = bookA.BestAsk().TotalVolume
				}
				// if symbol == "SOLUSD" && (a.Name() == "Cucoin" || b.Name() == "Cucoin") {
				// 	log.Println("a>b",symbol, b.Name(), bestBid.Price, a.Name(), bestAsk.Price)
				// }
			}

			// ->10$
			feeBidChain := sttg.Fee[symbol][bestBid.Provider]
			crossSpread.Spread = util.Round((bestAsk.Price*((10/bestBid.Price)-feeBidChain))-10, 10_0000)

			// ->100$
			// feeBidChain := sttg.Fee[symbol][bestBid.Provider]
			// crossSpread.Spread = util.Round((bestAsk.Price*((100/bestBid.Price)-feeBidChain))-100, 10_0000)

			// crossSpread.Spread = util.Round(bestAsk.Price-bestBid.Price, 10_0000)
			crossSpread.BestAsk = bestAsk
			crossSpread.BestBid = bestBid
			crossSpreads = append(crossSpreads, crossSpread)

			// if len(pvrs) == 2 {
			// 	continue
			// } else if j < len(pvrs)-2 {
			// 	j++
			// } else {
			// 	i++
			// 	j = i
			// }
			// if len(pvrs) == 2 {
			// 		break
			// }
			if j < len(pvrs)-1 {
				j++
			} else {
				i++
				j = i + 1
			}

		}
		// fmt.Printf("%v -> %+v",symbol, data[symbol])
		data[symbol] = crossSpreads
	}
	datach <- data
}
