//go:build !ratls
// +build !ratls

package main

import "github.com/gin-gonic/gin"

func registerRatlsRoutes(_ *gin.Engine) {
	// ratls routes are disabled in default build
}
