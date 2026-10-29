// Package handlerinterface declares the HTTP handler contracts the router uses.
package handlerinterface

import "github.com/gin-gonic/gin"

type HealthHandler interface {
	Health(c *gin.Context)
}
