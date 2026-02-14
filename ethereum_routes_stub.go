//go:build !ethereum
// +build !ethereum

package main

import "github.com/gin-gonic/gin"

func registerEthereumRoutes(_ *gin.Engine) {
	// ethereum routes are disabled in default build
}
