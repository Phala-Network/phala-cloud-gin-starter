//go:build ethereum
// +build ethereum

package main

import (
	"context"
	"net/http"

	"github.com/Dstack-TEE/dstack/sdk/go/dstack"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

func registerEthereumRoutes(r *gin.Engine) {
	// v1: GET /v1/ethereum?domain=dstack (dstack OS >= 0.6.0)
	r.GET("/v1/ethereum", func(c *gin.Context) {
		domain := c.DefaultQuery("domain", "dstack")
		client := dstack.NewDstackClient()
		resp, err := client.GetKey(context.Background(), domain, "secp256k1")
		if err != nil {
			recordFailure("v1_ethereum", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		priv, err := crypto.ToECDSA(resp.Key)
		if err != nil {
			recordFailure("v1_ethereum", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		recordSuccess()
		c.JSON(http.StatusOK, gin.H{"address": crypto.PubkeyToAddress(priv.PublicKey).Hex()})
	})

	// Legacy v0: keys differ from v1 for the same name.
	r.GET("/ethereum", func(c *gin.Context) {
		key := c.DefaultQuery("key", "dstack")
		client := dstack.NewDstackClientV0()
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
