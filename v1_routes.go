package main

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/Dstack-TEE/dstack/sdk/go/dstack"
	"github.com/gin-gonic/gin"
)

// registerV1Routes serves the dstack guest API v1 (dstack OS >= 0.6.0).
func registerV1Routes(r *gin.Engine) {
	g := r.Group("/v1")

	g.GET("/info", func(c *gin.Context) {
		client := dstack.NewDstackClient()
		resp, err := client.Info(context.Background())
		if err != nil {
			recordFailure("v1_info", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{
			"app_id":        hex.EncodeToString(resp.AppID),
			"app_name":      resp.AppName,
			"compose_hash":  hex.EncodeToString(resp.ComposeHash),
			"instance_id":   hex.EncodeToString(resp.InstanceID),
			"device_id":     hex.EncodeToString(resp.DeviceID),
			"os_image_hash": hex.EncodeToString(resp.OsImageHash),
			"mr_aggregated": hex.EncodeToString(resp.MrAggregated),
			"vm_config":     resp.VmConfig,
			"key_provider":  resp.KeyProviderInfo,
			"cloud_vendor":  resp.CloudVendor,
			"cloud_product": resp.CloudProduct,
			"app_compose":   resp.AppCompose,
		})
	})

	g.GET("/version", func(c *gin.Context) {
		client := dstack.NewDstackClient()
		resp, err := client.Version(context.Background())
		if err != nil {
			recordFailure("v1_version", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, resp)
	})

	// GET /v1/key?domain=dstack&algorithm=secp256k1
	// The private key is not returned; the public key and signature chain are
	// enough to show derivation.
	g.GET("/key", func(c *gin.Context) {
		domain := c.DefaultQuery("domain", "dstack")
		// v1 has no default algorithm; make the caller choose.
		algorithm := c.Query("algorithm")
		if algorithm != "secp256k1" && algorithm != "ed25519" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "algorithm must be secp256k1 or ed25519"})
			return
		}
		client := dstack.NewDstackClient()
		resp, err := client.GetKey(context.Background(), domain, algorithm)
		if err != nil {
			recordFailure("v1_key", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		chain := make([]string, len(resp.SignatureChain))
		for i, sig := range resp.SignatureChain {
			chain[i] = hex.EncodeToString(sig)
		}
		c.JSON(http.StatusOK, gin.H{
			"domain":          domain,
			"algorithm":       algorithm,
			"public_key":      hex.EncodeToString(resp.PublicKey),
			"signature_chain": chain,
			"root_epoch":      resp.RootEpoch,
			"app_epoch":       resp.AppEpoch,
		})
	})

	// GET /v1/attest?text=hello
	// v1 has a single attestation entry point; report data is 1-64 bytes.
	g.GET("/attest", func(c *gin.Context) {
		text := c.DefaultQuery("text", "hello dstack")
		reportData := appDataReportData([]byte(text))
		client := dstack.NewDstackClient()
		resp, err := client.Attest(context.Background(), reportData, false)
		if err != nil {
			recordFailure("v1_attest", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{
			"attestation":            hex.EncodeToString(resp.Attestation),
			"user_input":             text,
			"report_data_tag":        reportDataTag,
			"report_data_hex":        hex.EncodeToString(reportData),
			"report_data_derivation": reportDataDerivation,
		})
	})

	// GET /v1/tls_key?subject=localhost
	// IssueCert replaces v0 GetTlsKey. The generated private key is not returned.
	g.GET("/tls_key", func(c *gin.Context) {
		subject := c.DefaultQuery("subject", "localhost")
		client := dstack.NewDstackClient()
		resp, err := client.IssueCert(context.Background(), dstack.WithCertSubject(subject))
		if err != nil {
			recordFailure("v1_tls_key", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"subject": subject, "certificate_chain": resp.CertificateChain})
	})
}
