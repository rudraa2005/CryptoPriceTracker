# CryptoPriceTracker

CryptoPriceTracker is a small Go web app that proxies live CoinGecko data and serves a minimal frontend for:

- looking up a single crypto price in USD
- comparing two coins side by side
- checking the current trending coins list

The UI is embedded into the Go binary, so the project runs as a single server.

## Run locally

```bash
PORT=8080 API_KEY=your_key_here go run .
```

`API_KEY` is optional for public CoinGecko access, but useful if you have stricter rate limits or a paid plan.

Open `http://localhost:8080` and use CoinGecko coin IDs such as `bitcoin`, `ethereum`, or `solana`.

## API routes

- `GET /api/health`
- `GET /api/price?symbol=bitcoin`
- `GET /api/prices?symbol1=bitcoin&symbol2=ethereum`
- `GET /api/trending`
- `GET /api/supported-currencies`
