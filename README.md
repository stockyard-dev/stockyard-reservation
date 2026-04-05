# Stockyard Reservation

**Self-hosted reservation and table management**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/reservation/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9806:9806 -v reservation_data:/data ghcr.io/stockyard-dev/stockyard-reservation
```

Open `http://localhost:9806` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9806` | HTTP port |
| `DATA_DIR` | `./reservation-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
