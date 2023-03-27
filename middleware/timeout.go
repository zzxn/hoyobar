package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// make gin.Context.Request.Context() be with timeout
// users should use Request.Context() to leverage Timeout
// if gin engine's ContextWithFallback == true, then *gin.Context can be used too
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next() // must it call it to make cancel func called after all work done
	}
}
