package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)
type TrendingResponse struct {
	Coins []struct {
		Item struct {
			ID            string  `json:"id"`
			CoinID        int     `json:"coin_id"`
			Name          string  `json:"name"`
			Symbol        string  `json:"symbol"`
			MarketCapRank int     `json:"market_cap_rank"`
			Thumb         string  `json:"thumb"`
			PriceBTC      float64 `json:"price_btc"`
			Score         int     `json:"score"`
			Data          struct {
				Price                   float64            `json:"price"`
				PriceBTC                string             `json:"price_btc"`
				PriceChangePercentage24 map[string]float64 `json:"price_change_percentage_24h"`
				MarketCap               string             `json:"market_cap"`
				MarketCapBTC            string             `json:"market_cap_btc"`
				TotalVolume             string             `json:"total_volume"`
				TotalVolumeBTC          string             `json:"total_volume_btc"`
				Sparkline               string             `json:"sparkline"`
			} `json:"data"`
		} `json:"item"`
	} `json:"coins"`
	NFTs       []interface{} `json:"nfts"`
	Categories []interface{} `json:"categories"`
}
type APIError struct {
	Status struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}
func GetPrice(symbol string) float64 {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return -1
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("API Response:", string(body))
	if err := json.Unmarshal(body, &result); err != nil {
		return -1
	}
	return result[strings.ToLower(symbol)]["usd"]
}
func GetSupportedCoins() []string {
	url := "https://api.coingecko.com/api/v3/simple/supported_vs_currencies"
	req, _ := http.NewRequest("GET", url, nil)
	apiKey := os.Getenv("API_KEY")
	req.Header.Add("x-cg-pro-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result []string
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("API Response:", string(body))
	fmt.Println("API Response:", string(body))
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}
	return result
}

func GetPrices(symbol1 string, symbol2 string) (float64, float64) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s,%s&vs_currencies=usd", symbol1, symbol2)

	resp, err := http.Get(url)
	if err != nil {
		return -1, -1
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("API Response:", string(body))
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, -1
	}
	return result[strings.ToLower(symbol1)]["usd"], result[strings.ToLower(symbol2)]["usd"]
}

func GetTrending() (*TrendingResponse, error) {
	url := "https://api.coingecko.com/api/v3/search/trending"
	req, _ := http.NewRequest("GET", url, nil)
	apiKey := os.Getenv("API_KEY")
	req.Header.Add("x-cg-pro-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TrendingResponse
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("API Response:", string(body))
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

