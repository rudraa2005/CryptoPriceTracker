package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yourusername/crypto-price-tracker/internal/crypto"
)

func priceHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "missing symbol!", http.StatusBadRequest)
		return
	}
	price := crypto.GetPrice(symbol)

	resp := map[string]interface{}{
		"symbol": symbol,
		"price":  price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func pricesHandler(w http.ResponseWriter, r *http.Request) {
	symbol1 := r.URL.Query().Get("symbol1")
	symbol2 := r.URL.Query().Get("symbol2")
	if symbol1 == "" {
		http.Error(w, "missing symbol!", http.StatusBadRequest)
		return
	}
	if symbol2 == "" {
		http.Error(w, "missing symbol!", http.StatusBadRequest)
		return
	}
	price1, price2 := crypto.GetPrices(symbol1, symbol2)
	resp := map[string]interface{}{
		"symbol1": symbol1,
		"price1":  price1,
		"symbol2": symbol2,
		"price2":  price2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func supportedCoinsHandler(w http.ResponseWriter, r *http.Request) {
	coins := crypto.GetSupportedCoins()
	if coins == nil {
		http.Error(w, "failed to fetch supported coins", http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"coins": coins,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func trendingHandler(w http.ResponseWriter, r *http.Request) {
	trending_coins, err := crypto.GetTrending()
	if err != nil {
		http.Error(w, "Failed to fetch!", http.StatusInternalServerError)
		return
	}

	var coins []map[string]interface{}
	for _, c := range trending_coins.Coins {
		coins = append(coins, map[string]interface{}{
			"id":     c.Item.ID,
			"name":   c.Item.Name,
			"symbol": c.Item.Symbol,
			"rank":   c.Item.MarketCapRank,
			"price":  c.Item.Data.Price,
		})
	}
	resp := map[string]interface{}{
		"trending coins": coins,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🚀 Crypto Price Tracker API is running"))
	})

	// Price endpoint
	r.Get("/price", priceHandler)
	r.Get("/supported_coins", supportedCoinsHandler)
	r.Get("/prices", pricesHandler)
	r.Get("/trending", trendingHandler)

	// Port setup
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	http.ListenAndServe(":"+port, r)

}
