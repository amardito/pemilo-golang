package middleware

import (
	"net/http"

	"pemilo/config"

	"github.com/gin-gonic/gin"
)

const (
	SessionName   = "pemilo_session"
	UserIDKey     = "user_id"
	UsernameKey   = "username"
	ContextUserID = "userID"
)

// AuthRequired middleware checks if user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := config.SessionStore.Get(c.Request, SessionName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid session",
			})
			c.Abort()
			return
		}

		userID, ok := session.Values[UserIDKey].(uint)
		if !ok || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Set user ID in context for handlers
		c.Set(ContextUserID, userID)
		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(ContextUserID); exists {
		return userID.(uint)
	}
	return 0
}
