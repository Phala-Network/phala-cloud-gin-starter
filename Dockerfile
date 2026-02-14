FROM golang:1.24-alpine AS builder
WORKDIR /app

ARG GO_BUILD_TAGS=""

# Install build dependencies needed for CGO (dcap profile).
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# When building with ethereum tag, add optional dependency explicitly.
RUN if echo " $GO_BUILD_TAGS " | grep -q " ethereum "; then go get github.com/ethereum/go-ethereum@v1.16.8; fi

# Build: use CGO when dcap tag is present (expects libdcap_qvl.a in lib/), otherwise static build.
RUN if echo " $GO_BUILD_TAGS " | grep -q " dcap "; then \
      CGO_ENABLED=1 CGO_LDFLAGS="-L/app/lib" GOOS=linux GOARCH=amd64 \
        go build -tags "$GO_BUILD_TAGS" -o server . ; \
    else \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags "$GO_BUILD_TAGS" -o server . ; \
    fi

FROM alpine:3.21
WORKDIR /app
RUN adduser -D -u 10001 appuser
COPY --from=builder /app/server /app/server
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/server"]
