package coins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ResponseCoins struct {
	SymbolList []*ListData `json:"symbol_list"`
	
}

type ListData struct {
	
}

var urls = []string{
	"https://openapi.digifinex.com/v3/spot/symbols",
}

func UpdateCoins() {
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("error response %v: %v", url, err)
		}
		defer resp.Body.Close()

		// b ,_ :=io.ReadAll(resp.Body)
		data := ResponseCoins{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil{
			log.Printf("error Unmarshal response %v: %v", url, err)
		}

		fmt.Print(data)
	}

}
