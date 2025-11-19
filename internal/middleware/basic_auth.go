package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BasicAuth provides HTTP Basic Authentication for owner-only endpoints
func BasicAuth(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		// Extract Basic auth credentials
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Decode base64 credentials
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization encoding"})
			c.Abort()
			return
		}

		// Split username:password
		creds := strings.SplitN(string(decoded), ":", 2)
		if len(creds) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials format"})
			c.Abort()
			return
		}

		// Verify credentials
		if creds[0] != username || creds[1] != password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			c.Abort()
			return
		}

		c.Next()
	}
}
