//go:build ethereum
// +build ethereum

package main

import (
	"context"
	"net/http"

	"github.com/Dstack-TEE/dstack/sdk/go/dstack"
	"github.com/gin-gonic/gin"
)

func registerEthereumRoutes(r *gin.Engine) {
	r.GET("/ethereum", func(c *gin.Context) {
		key := c.DefaultQuery("key", "dstack")
		client := dstack.NewDstackClient()
		resp, err := client.GetKey(context.Background(), key, "", "secp256k1")
		if err != nil {
			recordFailure("ethereum", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		acct, err := dstack.ToEthereumAccountSecure(resp)
		if err != nil {
			recordFailure("ethereum", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"address": acct.Address.Hex()})
	})
}
