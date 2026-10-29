// Package middlewareinterface declares the middleware contracts the router uses.
package middlewareinterface

import "github.com/gin-gonic/gin"

type ErrorHandler interface {
	// Handle converts errors attached with c.Error into the standard JSON error response.
	Handle() gin.HandlerFunc
	// Recover turns panics into a 500 with the standard error body.
	Recover() gin.HandlerFunc
	NotFound() gin.HandlerFunc
	MethodNotAllowed() gin.HandlerFunc
}
