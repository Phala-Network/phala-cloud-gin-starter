//go:build ratls
// +build ratls

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Dstack-TEE/dstack/sdk/go/ratls"
	"github.com/gin-gonic/gin"
)

const ratlsDialTimeout = 15 * time.Second

func registerRatlsRoutes(r *gin.Engine) {
	g := r.Group("/ratls")

	// GET /ratls/verify?endpoint=host[:port][&pccs_url=...]
	// Connects to the given TLS endpoint, extracts the peer certificate,
	// and verifies it is a valid RA-TLS certificate with a genuine TDX quote.
	// Port defaults to 443 if not specified.
	g.GET("/verify", func(c *gin.Context) {
		endpoint := c.Query("endpoint")
		if endpoint == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'endpoint' query parameter"})
			return
		}
		endpoint = normalizeEndpoint(endpoint)

		pccsURL := c.DefaultQuery("pccs_url", ratls.DefaultPCCSURL)

		cert, err := fetchPeerCert(c.Request.Context(), endpoint)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":    "failed to connect: " + err.Error(),
				"endpoint": endpoint,
			})
			return
		}

		result, err := ratls.VerifyCert(cert, ratls.WithPCCSURL(pccsURL))
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":    err.Error(),
				"endpoint": endpoint,
				"subject":  cert.Subject.String(),
			})
			return
		}

		resp := gin.H{
			"verified":     true,
			"endpoint":     endpoint,
			"status":       result.Report.Status,
			"advisory_ids": result.Report.AdvisoryIDs,
			"quote_type":   result.Quote.QuoteType,
			"report_type":  result.Quote.Report.Type,
		}
		if len(result.Quote.Report.RTMR0) > 0 {
			resp["rtmr0"] = hex.EncodeToString(result.Quote.Report.RTMR0)
			resp["rtmr1"] = hex.EncodeToString(result.Quote.Report.RTMR1)
			resp["rtmr2"] = hex.EncodeToString(result.Quote.Report.RTMR2)
			resp["rtmr3"] = hex.EncodeToString(result.Quote.Report.RTMR3)
		}
		c.JSON(http.StatusOK, resp)
	})

	// GET /ratls/cert?endpoint=host[:port]
	// Connects to the TLS endpoint and returns certificate info without
	// performing RA-TLS verification. Useful for debugging.
	// Port defaults to 443 if not specified.
	g.GET("/cert", func(c *gin.Context) {
		endpoint := c.Query("endpoint")
		if endpoint == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'endpoint' query parameter"})
			return
		}
		endpoint = normalizeEndpoint(endpoint)

		cert, err := fetchPeerCert(c.Request.Context(), endpoint)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":    "failed to connect: " + err.Error(),
				"endpoint": endpoint,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"endpoint":   endpoint,
			"subject":    cert.Subject.String(),
			"issuer":     cert.Issuer.String(),
			"not_before": cert.NotBefore,
			"not_after":  cert.NotAfter,
			"extensions": describeExtensions(cert),
		})
	})
}

// normalizeEndpoint appends port 443 if no port is specified.
func normalizeEndpoint(endpoint string) string {
	_, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return net.JoinHostPort(endpoint, "443")
	}
	return endpoint
}

// fetchPeerCert connects to a TLS endpoint and returns the peer certificate.
func fetchPeerCert(parent context.Context, endpoint string) (*x509.Certificate, error) {
	ctx, cancel := context.WithTimeout(parent, ratlsDialTimeout)
	defer cancel()

	dialer := &tls.Dialer{
		Config: &tls.Config{InsecureSkipVerify: true},
	}
	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	state := conn.(*tls.Conn).ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("server presented no certificate")
	}
	return state.PeerCertificates[0], nil
}

// describeExtensions returns a summary of certificate extensions,
// labeling known Phala RA-TLS OIDs for easy identification.
func describeExtensions(cert *x509.Certificate) []gin.H {
	knownOIDs := map[string]string{
		"1.3.6.1.4.1.62397.1.1": "phala-ratls-tdx-quote",
		"1.3.6.1.4.1.62397.1.2": "phala-ratls-event-log",
		"1.3.6.1.4.1.62397.1.3": "phala-ratls-app-id",
		"1.3.6.1.4.1.62397.1.4": "phala-ratls-cert-usage",
		"1.3.6.1.4.1.62397.1.8": "phala-ratls-attestation",
		"1.3.6.1.4.1.62397.1.9": "phala-ratls-app-info",
	}

	var exts []gin.H
	for _, ext := range cert.Extensions {
		entry := gin.H{
			"oid":      ext.Id.String(),
			"critical": ext.Critical,
			"size":     len(ext.Value),
		}
		if name, ok := knownOIDs[ext.Id.String()]; ok {
			entry["name"] = name
		}
		exts = append(exts, entry)
	}
	return exts
}
