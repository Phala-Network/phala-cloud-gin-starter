//go:build dcap
// +build dcap

package main

import (
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	dcap "github.com/Phala-Network/dcap-qvl/golang-bindings"
	"github.com/gin-gonic/gin"
)

// collateralCache is a simple in-memory cache for collateral results keyed by "fmspc:quoteType".
var collateralCache sync.Map

type collateralEntry struct {
	data      *dcap.QuoteCollateralV3
	expiresAt time.Time
}

const collateralCacheTTL = 24 * time.Hour

func getCachedCollateral(key string) (*dcap.QuoteCollateralV3, bool) {
	v, ok := collateralCache.Load(key)
	if !ok {
		return nil, false
	}
	entry := v.(*collateralEntry)
	if time.Now().After(entry.expiresAt) {
		collateralCache.Delete(key)
		return nil, false
	}
	return entry.data, true
}

func setCachedCollateral(key string, coll *dcap.QuoteCollateralV3) {
	collateralCache.Store(key, &collateralEntry{
		data:      coll,
		expiresAt: time.Now().Add(collateralCacheTTL),
	})
}

// maxQuoteRequestSize limits the request body to 10 MiB to prevent DoS via unbounded reads.
const maxQuoteRequestSize = 10 << 20 // 10 MiB

// readQuoteFromRequest extracts raw quote bytes from the request.
// Supports: multipart file upload (field "file"), application/octet-stream body, JSON hex {"hex":"..."}
func readQuoteFromRequest(c *gin.Context) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxQuoteRequestSize)
	ct := c.ContentType()

	// multipart file upload
	if strings.HasPrefix(ct, "multipart/form-data") {
		f, err := c.FormFile("file")
		if err != nil {
			return nil, err
		}
		file, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(file)
	}

	// JSON hex string
	if strings.HasPrefix(ct, "application/json") {
		var body struct {
			Hex string `json:"hex"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			return nil, err
		}
		h := strings.TrimSpace(body.Hex)
		h = strings.TrimPrefix(h, "0x")
		h = strings.TrimPrefix(h, "0X")
		return hex.DecodeString(h)
	}

	// raw binary body (application/octet-stream or fallback)
	return io.ReadAll(c.Request.Body)
}

func registerDcapRoutes(r *gin.Engine) {
	g := r.Group("/dcap")

	// POST /dcap/parse - Parse a raw quote binary
	g.POST("/parse", func(c *gin.Context) {
		rawQuote, err := readQuoteFromRequest(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read quote: " + err.Error()})
			return
		}

		quote, err := dcap.ParseQuote(rawQuote)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to parse quote: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, quote)
	})

	// POST /dcap/verify - Parse, fetch collateral, and verify a quote
	g.POST("/verify", func(c *gin.Context) {
		rawQuote, err := readQuoteFromRequest(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read quote: " + err.Error()})
			return
		}

		report, err := dcap.GetCollateralAndVerify(rawQuote, dcap.PhalaPCCSURL)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "verification failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, report)
	})

	// POST /dcap/collateral - Fetch collateral for a quote
	g.POST("/collateral", func(c *gin.Context) {
		rawQuote, err := readQuoteFromRequest(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read quote: " + err.Error()})
			return
		}

		quote, err := dcap.ParseQuote(rawQuote)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to parse quote: " + err.Error()})
			return
		}

		cacheKey := quote.FMSPC + ":" + quote.QuoteType
		if coll, ok := getCachedCollateral(cacheKey); ok {
			c.JSON(http.StatusOK, coll)
			return
		}

		isSGX := quote.QuoteType == "SGX"
		coll, err := dcap.GetCollateralForFMSPC(dcap.PhalaPCCSURL, quote.FMSPC, quote.CA, isSGX)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch collateral: " + err.Error()})
			return
		}

		setCachedCollateral(cacheKey, coll)
		c.JSON(http.StatusOK, coll)
	})

	// POST /dcap/pck - Parse PCK certificate extension from quote
	g.POST("/pck", func(c *gin.Context) {
		rawQuote, err := readQuoteFromRequest(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read quote: " + err.Error()})
			return
		}

		quote, err := dcap.ParseQuote(rawQuote)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to parse quote: " + err.Error()})
			return
		}

		if quote.CertChainPEM == "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "quote has no embedded certificate chain"})
			return
		}

		ext, err := dcap.ParsePCKExtensionFromPEM([]byte(quote.CertChainPEM))
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to parse PCK extension: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, ext)
	})
}
