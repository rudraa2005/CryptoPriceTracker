package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

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
