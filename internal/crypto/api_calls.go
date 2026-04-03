package crypto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.coingecko.com/api/v3"

var (
	ErrInvalidCoinID = errors.New("coin id is required")
	ErrCoinNotFound  = errors.New("coin not found")
)

type UpstreamError struct {
	StatusCode int
	Message    string
}

func (e *UpstreamError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("upstream request failed with status %d", e.StatusCode)
	}

	return fmt.Sprintf("upstream request failed: %s", e.Message)
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type TrendingCoin struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	Rank   int     `json:"rank"`
	Price  float64 `json:"price"`
}

type trendingResponse struct {
	Coins []struct {
		Item struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Symbol        string `json:"symbol"`
			MarketCapRank int    `json:"market_cap_rank"`
			Data          struct {
				Price float64 `json:"price"`
			} `json:"data"`
		} `json:"item"`
	} `json:"coins"`
}

type apiErrorResponse struct {
	Status struct {
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  strings.TrimSpace(os.Getenv("API_KEY")),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetPrice(ctx context.Context, coinID string) (float64, error) {
	normalized, err := normalizeCoinID(coinID)
	if err != nil {
		return 0, err
	}

	prices, err := c.GetPrices(ctx, normalized)
	if err != nil {
		return 0, err
	}

	return prices[normalized], nil
}

func (c *Client) GetPrices(ctx context.Context, coinIDs ...string) (map[string]float64, error) {
	if len(coinIDs) == 0 {
		return nil, ErrInvalidCoinID
	}

	normalizedIDs := make([]string, 0, len(coinIDs))
	seen := make(map[string]struct{}, len(coinIDs))

	for _, coinID := range coinIDs {
		normalized, err := normalizeCoinID(coinID)
		if err != nil {
			return nil, err
		}

		if _, exists := seen[normalized]; exists {
			continue
		}

		seen[normalized] = struct{}{}
		normalizedIDs = append(normalizedIDs, normalized)
	}

	query := url.Values{}
	query.Set("ids", strings.Join(normalizedIDs, ","))
	query.Set("vs_currencies", "usd")

	var result map[string]map[string]float64
	if err := c.getJSON(ctx, "/simple/price", query, &result); err != nil {
		return nil, err
	}

	prices := make(map[string]float64, len(normalizedIDs))
	for _, coinID := range normalizedIDs {
		quote, ok := result[coinID]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrCoinNotFound, coinID)
		}

		price, ok := quote["usd"]
		if !ok {
			return nil, fmt.Errorf("usd price unavailable for %s", coinID)
		}

		prices[coinID] = price
	}

	return prices, nil
}

func (c *Client) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	var result []string
	if err := c.getJSON(ctx, "/simple/supported_vs_currencies", nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetTrending(ctx context.Context) ([]TrendingCoin, error) {
	var result trendingResponse
	if err := c.getJSON(ctx, "/search/trending", nil, &result); err != nil {
		return nil, err
	}

	coins := make([]TrendingCoin, 0, len(result.Coins))
	for _, coin := range result.Coins {
		item := coin.Item
		coins = append(coins, TrendingCoin{
			ID:     item.ID,
			Name:   item.Name,
			Symbol: strings.ToUpper(item.Symbol),
			Rank:   item.MarketCapRank,
			Price:  item.Data.Price,
		})
	}

	return coins, nil
}

func normalizeCoinID(coinID string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(coinID))
	if normalized == "" {
		return "", ErrInvalidCoinID
	}

	return normalized, nil
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values, dest any) error {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-cg-pro-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request upstream: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read upstream response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return parseUpstreamError(resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode upstream response: %w", err)
	}

	return nil
}

func parseUpstreamError(statusCode int, body []byte) error {
	message := strings.TrimSpace(string(body))
	var payload apiErrorResponse

	if len(body) > 0 && json.Unmarshal(body, &payload) == nil {
		switch {
		case payload.Status.ErrorMessage != "":
			message = payload.Status.ErrorMessage
		case payload.Error != "":
			message = payload.Error
		case payload.Message != "":
			message = payload.Message
		}
	}

	if message == "" {
		message = http.StatusText(statusCode)
	}

	return &UpstreamError{
		StatusCode: statusCode,
		Message:    message,
	}
}
