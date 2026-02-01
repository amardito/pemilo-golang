package routes

import (
	"pemilo/handlers"
	"pemilo/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine) {
	// Initialize WebSocket hub
	handlers.InitWebSocket()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 group
	api := r.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", middleware.RateLimitLogin(), handlers.Login)
			auth.POST("/register", handlers.Register) // TODO: Restrict in production
			auth.POST("/logout", handlers.Logout)
		}

		// Public voting routes (for voters)
		voting := api.Group("/vote")
		{
			voting.GET("/validate", handlers.ValidateTicket)
			voting.POST("/cast", handlers.CastVote)
		}

		// WebSocket for real-time updates
		api.GET("/ws/room/:roomId", handlers.HandleWebSocket)

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// User routes
			protected.GET("/me", handlers.GetMe)

			// Room routes
			rooms := protected.Group("/rooms")
			{
				rooms.GET("", handlers.ListRooms)
				rooms.POST("", handlers.CreateRoom)
				rooms.GET("/:id", handlers.GetRoom)
				rooms.PUT("/:id", handlers.UpdateRoom)
				rooms.DELETE("/:id", handlers.DeleteRoom)
				rooms.GET("/:id/results", handlers.GetRoomResults)
			}

			// Candidate routes (nested under rooms)
			candidates := protected.Group("/rooms/:roomId/candidates")
			{
				candidates.GET("", handlers.ListCandidates)
				candidates.POST("", handlers.CreateCandidate)
				candidates.GET("/:id", handlers.GetCandidate)
				candidates.PUT("/:id", handlers.UpdateCandidate)
				candidates.DELETE("/:id", handlers.DeleteCandidate)
			}

			// Ticket routes (nested under rooms)
			tickets := protected.Group("/rooms/:roomId/tickets")
			{
				tickets.GET("", handlers.ListTickets)
				tickets.POST("/generate", handlers.GenerateTickets)
				tickets.GET("/export", handlers.ExportTicketsCSV)
				tickets.DELETE("/:id", handlers.DeleteTicket)
			}
		}
	}
}
