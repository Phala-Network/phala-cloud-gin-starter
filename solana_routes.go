//go:build solana
// +build solana

package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/Dstack-TEE/dstack/sdk/go/dstack"
	"github.com/gin-gonic/gin"
)

func registerSolanaRoutes(r *gin.Engine) {
	// v1: GET /v1/solana?domain=dstack (dstack OS >= 0.6.0)
	// The v1 key is a 32-byte ed25519 seed. The public key is returned as hex;
	// base58 encoding is left out to avoid an extra dependency.
	r.GET("/v1/solana", func(c *gin.Context) {
		domain := c.DefaultQuery("domain", "dstack")
		client := dstack.NewDstackClient()
		resp, err := client.GetKey(context.Background(), domain, "ed25519")
		if err != nil {
			recordFailure("v1_solana", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if len(resp.Key) != ed25519.SeedSize {
			err := fmt.Errorf("unexpected ed25519 key length %d", len(resp.Key))
			recordFailure("v1_solana", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		pub := ed25519.NewKeyFromSeed(resp.Key).Public().(ed25519.PublicKey)
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"public_key": hex.EncodeToString(pub)})
	})

	// Legacy v0: keys differ from v1 for the same name.
	r.GET("/solana", func(c *gin.Context) {
		key := c.DefaultQuery("key", "dstack")
		client := dstack.NewDstackClientV0()
		resp, err := client.GetKey(context.Background(), key, "", "ed25519")
		if err != nil {
			recordFailure("solana", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		kp, err := dstack.ToSolanaKeypairSecure(resp)
		if err != nil {
			recordFailure("solana", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"address": kp.PublicKeyString()})
	})
}
