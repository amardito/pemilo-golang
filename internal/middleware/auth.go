package middleware

import (
	"net/http"
	"strings"

	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "invalid authorization format"})
			return
		}

		userID, err := service.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "invalid or expired token"})
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID extracts user ID from gin context.
func GetUserID(c *gin.Context) string {
	v, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	return v.(string)
}
