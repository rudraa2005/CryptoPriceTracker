# CryptoPriceTracker

CryptoPriceTracker is a small Go web app that proxies live CoinGecko data and serves a minimal frontend for:

- looking up a single crypto price in USD
- comparing two coins side by side
- checking the current trending coins list

The UI is embedded into the Go binary, so the project runs as a single server.

## Features

- minimal black, white, and grey interface with no frontend build step
- embedded static frontend served directly by the Go server
- JSON API for prices, comparisons, trending coins, and health checks
- CoinGecko-backed market data with optional API key support

## Requirements

- Go 1.24.6 or newer

## Configuration

The app reads configuration from environment variables and also auto-loads a local `.env` file at startup if it exists.

- `PORT`: HTTP port for the local server. Defaults to `8080`.
- `API_KEY`: optional CoinGecko API key. Useful if you want better rate limits or use a paid CoinGecko plan.

Example:

```bash
API_KEY=your_coingecko_key
PORT=8080
```

Save that into a local `.env` file, then run:

```bash
go run .
```

## Run locally

Open `http://localhost:8080` and use CoinGecko coin IDs such as `bitcoin`, `ethereum`, or `solana`.

## Frontend

The root route `/` serves the web interface. The page includes:

- a single-coin lookup panel
- a two-coin comparison panel
- a trending coins panel with manual refresh

## API routes

- `GET /api/health`
- `GET /api/price?symbol=bitcoin`
- `GET /api/prices?symbol1=bitcoin&symbol2=ethereum`
- `GET /api/trending`
- `GET /api/supported-currencies`

Backward-compatible aliases from the earlier API shape are also available:

- `GET /price`
- `GET /prices`
- `GET /trending`
- `GET /supported_coins`

## Example requests

```bash
curl "http://localhost:8080/api/price?symbol=bitcoin"
curl "http://localhost:8080/api/prices?symbol1=bitcoin&symbol2=ethereum"
curl "http://localhost:8080/api/trending"
```

## Testing

```bash
go test ./...
go vet ./...
```
