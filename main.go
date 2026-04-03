package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rudraa2005/crypto-price-tracker/internal/crypto"
)

//go:embed web/*
var webFiles embed.FS

type cryptoService interface {
	GetPrice(ctx context.Context, coinID string) (float64, error)
	GetPrices(ctx context.Context, coinIDs ...string) (map[string]float64, error)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)
	GetTrending(ctx context.Context) ([]crypto.TrendingCoin, error)
}

type server struct {
	service cryptoService
	static  http.Handler
}

func newServer(service cryptoService) *server {
	staticFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}

	return &server{
		service: service,
		static:  http.FileServer(http.FS(staticFS)),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/price", s.priceHandler)
	mux.HandleFunc("/api/prices", s.pricesHandler)
	mux.HandleFunc("/api/supported-currencies", s.supportedCurrenciesHandler)
	mux.HandleFunc("/api/trending", s.trendingHandler)

	// Backward-compatible aliases for the original API shape.
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/price", s.priceHandler)
	mux.HandleFunc("/prices", s.pricesHandler)
	mux.HandleFunc("/supported_coins", s.supportedCurrenciesHandler)
	mux.HandleFunc("/trending", s.trendingHandler)

	mux.Handle("/", s.static)

	return recoverMiddleware(loggingMiddleware(mux))
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) priceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol := normalizeSymbol(r.URL.Query().Get("symbol"))
	if symbol == "" {
		writeError(w, http.StatusBadRequest, "missing symbol query parameter")
		return
	}

	price, err := s.service.GetPrice(r.Context(), symbol)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"symbol": symbol,
		"price":  price,
	})
}

func (s *server) pricesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol1 := normalizeSymbol(r.URL.Query().Get("symbol1"))
	symbol2 := normalizeSymbol(r.URL.Query().Get("symbol2"))

	if symbol1 == "" || symbol2 == "" {
		writeError(w, http.StatusBadRequest, "both symbol1 and symbol2 query parameters are required")
		return
	}

	prices, err := s.service.GetPrices(r.Context(), symbol1, symbol2)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"symbol1": symbol1,
		"price1":  prices[symbol1],
		"symbol2": symbol2,
		"price2":  prices[symbol2],
	})
}

func (s *server) supportedCurrenciesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currencies, err := s.service.GetSupportedCurrencies(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"currencies": currencies,
	})
}

func (s *server) trendingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	coins, err := s.service.GetTrending(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"coins": coins,
	})
}

func normalizeSymbol(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func handleServiceError(w http.ResponseWriter, err error) {
	var upstreamErr *crypto.UpstreamError

	switch {
	case errors.Is(err, crypto.ErrInvalidCoinID):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, crypto.ErrCoinNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.As(err, &upstreamErr):
		writeError(w, http.StatusBadGateway, upstreamErr.Error())
	default:
		writeError(w, http.StatusBadGateway, "failed to fetch crypto data")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(writer, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, writer.status, time.Since(start).Round(time.Millisecond))
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func main() {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	srv := newServer(crypto.NewClient())

	log.Printf("server running on port %s", port)
	if err := http.ListenAndServe(":"+port, srv.routes()); err != nil {
		log.Fatal(err)
	}
}
