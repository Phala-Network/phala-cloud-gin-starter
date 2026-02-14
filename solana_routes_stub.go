//go:build !solana
// +build !solana

package main

import "github.com/gin-gonic/gin"

func registerSolanaRoutes(_ *gin.Engine) {
	// solana routes are disabled in default build
}
