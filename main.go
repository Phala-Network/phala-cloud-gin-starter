package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Dstack-TEE/dstack/sdk/go/dstack"
	"github.com/gin-gonic/gin"
)

const reportDataTag = "app-data"
const reportDataDerivation = `sha512("app-data:" || utf8(user_input))`

func appDataReportData(input []byte) []byte {
	h := sha512.New()
	h.Write([]byte(reportDataTag))
	h.Write([]byte(":"))
	h.Write(input)
	return h.Sum(nil)
}

type quoteDemoResponse struct {
	*dstack.GetQuoteResponse
	UserInput            string `json:"user_input"`
	ReportDataTag        string `json:"report_data_tag"`
	ReportDataHex        string `json:"report_data_hex"`
	ReportDataDerivation string `json:"report_data_derivation"`
}

func quoteDemo(input string, reportData []byte, resp *dstack.GetQuoteResponse) quoteDemoResponse {
	return quoteDemoResponse{
		GetQuoteResponse:     resp,
		UserInput:            input,
		ReportDataTag:        reportDataTag,
		ReportDataHex:        hex.EncodeToString(reportData),
		ReportDataDerivation: reportDataDerivation,
	}
}

var (
	failureThreshold    int32 = 10
	consecutiveFailures int32
)

func main() {
	if v := os.Getenv("FAILURE_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			failureThreshold = int32(n)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		client := dstack.NewDstackClient()
		resp, err := client.Info(context.Background())
		if err != nil {
			recordFailure("info", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/get_quote", func(c *gin.Context) {
		text := c.DefaultQuery("text", "hello dstack")
		reportData := appDataReportData([]byte(text))
		client := dstack.NewDstackClient()
		resp, err := client.GetQuote(context.Background(), reportData)
		if err != nil {
			recordFailure("get_quote", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, quoteDemo(text, reportData, resp))
	})

	r.GET("/tdx_quote", func(c *gin.Context) {
		text := c.DefaultQuery("text", "hello dstack")
		reportData := appDataReportData([]byte(text))
		client := dstack.NewDstackClient()
		resp, err := client.GetQuote(context.Background(), reportData)
		if err != nil {
			recordFailure("tdx_quote", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, quoteDemo(text, reportData, resp))
	})

	r.GET("/get_key", func(c *gin.Context) {
		key := c.DefaultQuery("key", "dstack")
		client := dstack.NewDstackClient()
		resp, err := client.GetKey(context.Background(), key, "", "secp256k1")
		if err != nil {
			recordFailure("get_key", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, resp)
	})

	registerEthereumRoutes(r)
	registerSolanaRoutes(r)
	registerDcapRoutes(r)
	registerRatlsRoutes(r)

	r.GET("/env", func(c *gin.Context) {
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"env": os.Environ()})
	})

	r.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		client := dstack.NewDstackClient()
		_, err := client.Info(ctx)
		if err != nil {
			recordFailure("healthz", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	log.Printf("listening on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func recordSuccess() {
	atomic.StoreInt32(&consecutiveFailures, 0)
}

func recordFailure(context string, err error) {
	count := atomic.AddInt32(&consecutiveFailures, 1)
	log.Printf("%s failed (%d/%d): %v", context, count, failureThreshold, err)
	if count >= failureThreshold {
		log.Printf("failure threshold reached, exiting to trigger restart")
		go func() {
			time.Sleep(50 * time.Millisecond)
			os.Exit(1)
		}()
	}
}
