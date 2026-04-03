package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudraa2005/crypto-price-tracker/internal/crypto"
)

type stubService struct {
	price      float64
	prices     map[string]float64
	currencies []string
	trending   []crypto.TrendingCoin
	err        error
}

func (s stubService) GetPrice(context.Context, string) (float64, error) {
	return s.price, s.err
}

func (s stubService) GetPrices(context.Context, ...string) (map[string]float64, error) {
	return s.prices, s.err
}

func (s stubService) GetSupportedCurrencies(context.Context) ([]string, error) {
	return s.currencies, s.err
}

func (s stubService) GetTrending(context.Context) ([]crypto.TrendingCoin, error) {
	return s.trending, s.err
}

func TestPriceHandlerRequiresSymbol(t *testing.T) {
	srv := newServer(stubService{})

	req := httptest.NewRequest(http.MethodGet, "/api/price", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "missing symbol") {
		t.Fatalf("expected error message, got %q", rec.Body.String())
	}
}

func TestPriceHandlerReturnsNormalizedSymbol(t *testing.T) {
	srv := newServer(stubService{price: 84235.12})

	req := httptest.NewRequest(http.MethodGet, "/api/price?symbol=Bitcoin", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"price"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Symbol != "bitcoin" {
		t.Fatalf("expected normalized symbol, got %q", payload.Symbol)
	}

	if payload.Price != 84235.12 {
		t.Fatalf("expected price 84235.12, got %f", payload.Price)
	}
}

func TestPriceHandlerMapsNotFoundError(t *testing.T) {
	srv := newServer(stubService{err: fmt.Errorf("%w: made-up-coin", crypto.ErrCoinNotFound)})

	req := httptest.NewRequest(http.MethodGet, "/api/price?symbol=made-up-coin", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRootServesFrontend(t *testing.T) {
	srv := newServer(stubService{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected html content type, got %q", contentType)
	}

	if !strings.Contains(rec.Body.String(), "Crypto prices without the noise.") {
		t.Fatalf("expected frontend content, got %q", rec.Body.String())
	}
}

func TestLoadDotEnvSetsValues(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")

	if err := os.WriteFile(envFile, []byte("TEST_API_KEY=test-key\nTEST_PORT=9090\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if err := loadDotEnv(envFile); err != nil {
		t.Fatalf("loadDotEnv returned error: %v", err)
	}

	if got := os.Getenv("TEST_API_KEY"); got != "test-key" {
		t.Fatalf("expected TEST_API_KEY to be set, got %q", got)
	}

	if got := os.Getenv("TEST_PORT"); got != "9090" {
		t.Fatalf("expected TEST_PORT to be set, got %q", got)
	}
}

func TestLoadDotEnvDoesNotOverrideExistingValues(t *testing.T) {
	t.Setenv("TEST_API_KEY", "existing-key")

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")

	if err := os.WriteFile(envFile, []byte("TEST_API_KEY=file-key\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if err := loadDotEnv(envFile); err != nil {
		t.Fatalf("loadDotEnv returned error: %v", err)
	}

	if got := os.Getenv("TEST_API_KEY"); got != "existing-key" {
		t.Fatalf("expected existing TEST_API_KEY to win, got %q", got)
	}
}
