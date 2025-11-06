package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger emits structured fields through the provided logger.
func RequestLogger(l *log.Logger) gin.HandlerFunc {
	if l == nil {
		l = log.Default()
	}

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		l.Printf("method=%s path=%s status=%d latency=%s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}
