package crypto

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newTestClient(handler roundTripFunc) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Transport: handler,
		},
	}
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

func TestGetPricesBuildsExpectedRequest(t *testing.T) {
	t.Parallel()

	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v3/simple/price" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if got := r.URL.Query().Get("ids"); got != "bitcoin,ethereum" {
			t.Fatalf("unexpected ids query: %s", got)
		}

		if got := r.URL.Query().Get("vs_currencies"); got != "usd" {
			t.Fatalf("unexpected vs_currencies query: %s", got)
		}

		return jsonResponse(http.StatusOK, `{"bitcoin":{"usd":92345.67},"ethereum":{"usd":1765.44}}`), nil
	})

	prices, err := client.GetPrices(context.Background(), "Bitcoin", "ethereum")
	if err != nil {
		t.Fatalf("GetPrices returned error: %v", err)
	}

	if got := prices["bitcoin"]; got != 92345.67 {
		t.Fatalf("unexpected bitcoin price: %f", got)
	}

	if got := prices["ethereum"]; got != 1765.44 {
		t.Fatalf("unexpected ethereum price: %f", got)
	}
}

func TestGetPricesReturnsCoinNotFound(t *testing.T) {
	t.Parallel()

	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"bitcoin":{"usd":92345.67}}`), nil
	})

	_, err := client.GetPrices(context.Background(), "bitcoin", "solana")
	if !errors.Is(err, ErrCoinNotFound) {
		t.Fatalf("expected ErrCoinNotFound, got %v", err)
	}
}

func TestGetTrendingTransformsResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v3/search/trending" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"coins": [
				{
					"item": {
						"id": "bitcoin",
						"name": "Bitcoin",
						"symbol": "btc",
						"market_cap_rank": 1,
						"data": {
							"price": 92345.67
						}
					}
				}
			]
		}`), nil
	})

	coins, err := client.GetTrending(context.Background())
	if err != nil {
		t.Fatalf("GetTrending returned error: %v", err)
	}

	if len(coins) != 1 {
		t.Fatalf("expected 1 coin, got %d", len(coins))
	}

	if coins[0].Symbol != "BTC" {
		t.Fatalf("expected symbol BTC, got %s", coins[0].Symbol)
	}

	if coins[0].Price != 92345.67 {
		t.Fatalf("unexpected price: %f", coins[0].Price)
	}
}

func TestParseUpstreamErrorUsesStructuredMessage(t *testing.T) {
	t.Parallel()

	err := parseUpstreamError(http.StatusTooManyRequests, []byte(`{"status":{"error_message":"rate limit"}}`))

	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T", err)
	}

	if upstreamErr.Message != "rate limit" {
		t.Fatalf("unexpected upstream message: %s", upstreamErr.Message)
	}
}
