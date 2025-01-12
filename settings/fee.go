package settings

var Fee = map[string]map[string]float64{
	"XLMUSD": {
		// "Finex":    "XLMUSDTPERP", // не поддерживается
		// "Kraken":   "XLM/USD", // не работает в России
		// "Coinbase": "XLM-USD", // не работает в России

		"Bybit":   0.02,
		"Binance": 0.02,
		"OKX":     0.016,
		"MEXC":    0.1,
		"Cucoin":  3,
		"Gateio":  4.25,
		"HTX":     0.02,
	},
	"WAXPUSD": {
		// "Kraken":   "WAX/USD", // не работает в России
		// "Coinbase": "WAXT-USD", // не поддерживается
		// "Finex":  "WAXPUSDTPERP", // не поддерживается

		"Bybit":   2,
		"Binance": 2,
		"OKX":     0.1,
		"MEXC":    2,
		"Cucoin":  9.158,
		"Gateio":  8.95,
		"HTX":     1.5,
	},
	"ALGOUSD": {
		// "Finex":    "SOLUSDTPERP", // высокие комиссии
		// "Kraken":   "SOL/USD", // не работает в России
		// "Coinbase": "SOL-USD", // не работает в России

		"Bybit":   0.01,
		"Binance": 0.008,
		"OKX":     0.008,
		"MEXC":    0.1,
		"Cucoin":  0.1,
		"Gateio":  2.94,
		"HTX":     0.01,
	},
}