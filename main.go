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

	// Port setup
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	http.ListenAndServe(":"+port, r)

}
