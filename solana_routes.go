//go:build solana
// +build solana

package main

import (
	"context"
	"net/http"

	"github.com/Dstack-TEE/dstack/sdk/go/dstack"
	"github.com/gin-gonic/gin"
)

func registerSolanaRoutes(r *gin.Engine) {
	r.GET("/solana", func(c *gin.Context) {
		key := c.DefaultQuery("key", "dstack")
		client := dstack.NewDstackClient()
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
