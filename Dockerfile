FROM golang:1.24-alpine AS builder
WORKDIR /app

ARG GO_BUILD_TAGS=""

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# When building with ethereum tag, add optional dependency explicitly.
RUN if echo " $GO_BUILD_TAGS " | grep -q " ethereum "; then go get github.com/ethereum/go-ethereum@v1.16.8; fi
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags "$GO_BUILD_TAGS" -o server .

FROM alpine:3.21
WORKDIR /app
RUN adduser -D -u 10001 appuser
COPY --from=builder /app/server /app/server
USER appuser
EXPOSE 3000
ENTRYPOINT ["/app/server"]
