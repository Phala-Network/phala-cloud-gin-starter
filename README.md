# Phala Cloud Gin + Go Starter

[![](https://cloud.phala.network/deploy-button.svg)](https://cloud.phala.network/)

A minimal [Gin](https://gin-gonic.com/) starter for Phala Cloud / dstack.

This starter tracks the latest Go SDK from `github.com/Dstack-TEE/dstack/sdk/go`.

If you want to test a local SDK branch, you can temporarily add:

```go
replace github.com/Dstack-TEE/dstack/sdk/go => ../dstack/sdk/go
```

## Build profiles

### 1) Core profile (default)

No optional blockchain helper routes.

Endpoints:

- `GET /`
- `GET /get_quote?text=hello`
- `GET /tdx_quote?text=hello`
- `GET /get_key?key=dstack`
- `GET /env`
- `GET /healthz`

Run:

```bash
go run .
```

### 2) Ethereum profile (`-tags ethereum`)

Adds route:

- `GET /ethereum?key=dstack`

Run:

```bash
go get github.com/ethereum/go-ethereum@v1.16.8
go run -tags ethereum .
```

### 3) Solana profile (`-tags solana`)

Adds route:

- `GET /solana?key=dstack`

Run:

```bash
go run -tags solana .
```

### 4) Both Ethereum + Solana

```bash
go get github.com/ethereum/go-ethereum@v1.16.8
go run -tags "ethereum solana" .
```

## Local development

```bash
cd /Users/leechael/workshop/phala/phala-cloud-gin-starter
go mod tidy
go run .
```

Default port is `:8080`.

Use simulator sockets:

```bash
export DSTACK_SIMULATOR_ENDPOINT=/absolute/path/to/dstack.sock
export TAPPD_SIMULATOR_ENDPOINT=/absolute/path/to/tappd.sock
```

## Docker build

Core profile:

```bash
docker build -t phala-cloud-gin-starter:local .
```

Ethereum profile:

```bash
docker build --build-arg GO_BUILD_TAGS=ethereum -t phala-cloud-gin-starter:ethereum .
```

Solana profile:

```bash
docker build --build-arg GO_BUILD_TAGS=solana -t phala-cloud-gin-starter:solana .
```

Both:

```bash
docker build --build-arg GO_BUILD_TAGS="ethereum solana" -t phala-cloud-gin-starter:full .
```

## Image tags (GHCR)

Workflow publishes profile-suffixed tags:

- `latest-core`
- `latest-ethereum`
- `latest-solana`
- `latest-full`
- `latest` (alias of full)

For tagged releases (e.g. `v0.1.0`):

- `v0.1.0-core`
- `v0.1.0-ethereum`
- `v0.1.0-solana`
- `v0.1.0-full`
- `v0.1.0` (alias of full)

SHA tags are also published with the same suffix pattern.

## Deploy

Use `docker-compose.yml` for Phala Cloud custom compose deployment.
