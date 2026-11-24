# Phala Cloud Gin + Go Starter

[![](https://cloud.phala.network/deploy-button.svg)](https://cloud.phala.network/)

A minimal [Gin](https://gin-gonic.com/) starter for Phala Cloud / dstack.

This starter demonstrates the dstack guest API v1 (`NewDstackClient`) from `github.com/Dstack-TEE/dstack/sdk/go/dstack` v0.6.1.

**Requirement:** the v1 routes need dstack OS >= 0.6.0. Phala Cloud CPU images `dstack-0.6.x` have it. GPU images are still 0.5.15 and only speak the legacy v0 API (see [Legacy v0 API](#legacy-v0-api)).

## v1 routes

- `GET /` lists the v1 and legacy routes
- `GET /v1/info`
- `GET /v1/version`
- `GET /v1/key?domain=dstack&algorithm=secp256k1` (`algorithm` is `secp256k1` or `ed25519`; returns the public key and signature chain, not the private key)
- `GET /v1/attest?text=hello` (report data is `sha512("app-data:" || utf8(text))`; v1 accepts 1-64 bytes)
- `GET /v1/tls_key?subject=localhost` (issues a certificate; the private key is not returned)
- `GET /v1/ethereum?domain=dstack` (ethereum profile)
- `GET /v1/solana?domain=dstack` (solana profile; returns the hex public key)

v1 has no `Sign`, `Verify`, `GetQuote` or `EmitEvent`; the attestation from `Attest` already carries the quote and event log.

Do not hold assets in a wallet derived from `GetKey`: the key is a function of the KMS root, so these wallets suit demos and low-value or rotatable signing only. See the [dstack SDK warning](https://github.com/Dstack-TEE/dstack/blob/next/sdk/js/README.md#blockchain-helpers).

## Legacy v0 API

The routes below use the deprecated v0 API (`NewDstackClientV0`) and work on every dstack OS version, including the 0.5.15 GPU images:

- `GET /info`, `GET /get_quote`, `GET /tdx_quote`, `GET /get_key`, `GET /ethereum`, `GET /solana`, and the DCAP / RA-TLS routes.

**Warning:** v1 `GetKey` derives different key material from v0 for the same name (v1 takes `domain` and a required `algorithm`; v0 takes `path` and `purpose`). An app that already published an address derived with v0 must stay on v0, or its address changes.

If you want to test a local SDK checkout, you can temporarily add:

```go
replace github.com/Dstack-TEE/dstack/sdk/go/dstack => ../dstack/sdk/go/dstack
```

## Build profiles

### 1) Core profile (default)

No optional blockchain helper routes.

Endpoints (legacy v0 routes; see above for v1):

- `GET /info`
- `GET /get_quote?text=hello`
- `GET /tdx_quote?text=hello`

Quote endpoints hash caller input as `sha512("app-data:" || utf8(text))` before
it becomes TDX `report_data`, so the quote cannot be reused as an RA-TLS client
certificate. The JSON includes `user_input`, `report_data_hex`, and
`report_data_derivation` so a verifier can recompute the digest.
- `GET /get_key?key=dstack`
- `GET /env`
- `GET /healthz`

Run:

```bash
go run .
```

### 2) Ethereum profile (`-tags ethereum`)

Adds routes:

- `GET /v1/ethereum?domain=dstack`
- `GET /ethereum?key=dstack` (legacy v0)

Run:

```bash
go get github.com/ethereum/go-ethereum@v1.17.10
go run -tags ethereum .
```

### 3) Solana profile (`-tags solana`)

Adds routes:

- `GET /v1/solana?domain=dstack`
- `GET /solana?key=dstack` (legacy v0)

Run:

```bash
go run -tags solana .
```

### 4) DCAP profile (`-tags dcap`)

Adds routes:

- `POST /dcap/parse` — Parse a raw quote binary
- `POST /dcap/verify` — Parse, fetch collateral, and verify a quote
- `POST /dcap/collateral` — Fetch collateral for a quote
- `POST /dcap/pck` — Extract PCK certificate extension from quote
- `GET /dcap/device_id` — Get device ID (SHA256 of PPID from verified TDX attestation)

Run:

```bash
go run -tags dcap .
```

### 5) Both Ethereum + Solana

```bash
go get github.com/ethereum/go-ethereum@v1.17.10
go run -tags "ethereum solana" .
```

## Local development

```bash
go mod tidy
go run .
```

Default port is `:8080`.

Local development needs a running dstack guest-agent simulator (Linux). See [sdk/simulator](https://github.com/Dstack-TEE/dstack/tree/next/sdk/simulator).

Point the SDK at it with `DSTACK_SIMULATOR_ENDPOINT` — a unix socket path, or an `http://` / `https://` URL (the v0 client accepts both):

```bash
export DSTACK_SIMULATOR_ENDPOINT=/path/to/dstack.sock
# or: export DSTACK_SIMULATOR_ENDPOINT=http://127.0.0.1:8090
go run .
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

DCAP profile:

```bash
docker build --build-arg GO_BUILD_TAGS=dcap -t phala-cloud-gin-starter:dcap .
```

Full (all profiles):

```bash
docker build --build-arg GO_BUILD_TAGS="ethereum solana dcap" -t phala-cloud-gin-starter:full .
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

## Release workflow

This repo supports two release triggers:

1. PR comment trigger:
   - Comment on a PR: `!release patch`
   - or `!release minor`
   - or `!release major`

2. Manual trigger:
   - Actions -> `release` workflow -> Run workflow
   - choose bump type: `patch|minor|major`

Release behavior:

- computes next semantic version tag: `vX.Y.Z`
- creates and pushes the git tag
- builds and publishes GHCR images (amd64):
  - `${version}-core|ethereum|solana|full`
  - `${version}` (alias of full)
  - `latest-*` and `latest` (full alias)
- creates GitHub Release with generated changelog

## Deploy

Use `docker-compose.yml` for Phala Cloud custom compose deployment.
