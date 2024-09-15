package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		<-ctx.Done()
		c.AbortWithStatus(408)
	}
}
