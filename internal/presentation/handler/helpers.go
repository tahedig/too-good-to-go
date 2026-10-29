package handler

import "github.com/gin-gonic/gin"

// abortWithError hands err to the global error middleware, which writes the
// standard error response. Handlers should return through this instead of
// calling c.JSON with an error body themselves.
//
// Example: if err != nil { abortWithError(c, err); return }
func abortWithError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}
