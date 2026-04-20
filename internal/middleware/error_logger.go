package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorLogger logs the full error details for any 5xx response.
// Handlers should call c.Error(err) before returning a 5xx so the cause is captured.
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		if status < http.StatusInternalServerError {
			return
		}

		ginErrs := c.Errors.Errors()
		if len(ginErrs) > 0 {
			log.Printf("[ERROR] %s %s → %d | %s",
				c.Request.Method,
				c.Request.URL.Path,
				status,
				strings.Join(ginErrs, "; "),
			)
		} else {
			log.Printf("[ERROR] %s %s → %d (no error attached — add c.Error(err) in handler)",
				c.Request.Method,
				c.Request.URL.Path,
				status,
			)
		}
	}
}
