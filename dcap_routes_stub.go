//go:build !dcap
// +build !dcap

package main

import "github.com/gin-gonic/gin"

func registerDcapRoutes(_ *gin.Engine) {
	// dcap routes are disabled in default build
}
