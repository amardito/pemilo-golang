package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/amardito/pemilo-golang/internal/config"
	"github.com/amardito/pemilo-golang/internal/handler"
	"github.com/amardito/pemilo-golang/internal/middleware"
	"github.com/amardito/pemilo-golang/internal/repository"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Initialize repositories
	roomRepo := repository.NewRoomRepository(db)
	candidateRepo := repository.NewCandidateRepository(db)
	subCandidateRepo := repository.NewSubCandidateRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)

	// Initialize usecases
	roomUsecase := usecase.NewRoomUsecase(roomRepo)
	candidateUsecase := usecase.NewCandidateUsecase(candidateRepo, subCandidateRepo, roomRepo)
	ticketUsecase := usecase.NewTicketUsecase(ticketRepo, roomRepo)
	votingUsecase := usecase.NewVotingUsecase(voteRepo, roomRepo, candidateRepo, ticketRepo)
	authUsecase := usecase.NewAuthUsecase(adminRepo, loginAttemptRepo, cfg.JWTSecret, cfg.EncryptionKey)
	adminUsecase := usecase.NewAdminUsecase(adminRepo, roomRepo, cfg.EncryptionKey)

	// Initialize handlers
	roomHandler := handler.NewRoomHandler(roomUsecase)
	candidateHandler := handler.NewCandidateHandler(candidateUsecase)
	ticketHandler := handler.NewTicketHandler(ticketUsecase)
	votingHandler := handler.NewVotingHandler(votingUsecase)
	authHandler := handler.NewAuthHandler(authUsecase)
	adminHandler := handler.NewAdminHandler(adminUsecase)

	// Setup router
	router := setupRouter(cfg, roomHandler, candidateHandler, ticketHandler, votingHandler, authHandler, adminHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(
	cfg *config.Config,
	roomHandler *handler.RoomHandler,
	candidateHandler *handler.CandidateHandler,
	ticketHandler *handler.TicketHandler,
	votingHandler *handler.VotingHandler,
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
) *gin.Engine {
	// Set gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// Owner routes (Basic Auth only)
		owner := v1.Group("/owner")
		owner.Use(middleware.BasicAuth(cfg.OwnerUsername, cfg.OwnerPassword))
		{
			owner.POST("/create-admin", adminHandler.CreateAdmin)
		}

		// Public voter routes
		vote := v1.Group("/vote")
		{
			vote.GET("", votingHandler.GetVoterRoomInfo)
			vote.POST("", votingHandler.CastVote)
			vote.POST("/verify-ticket", votingHandler.VerifyTicket)
		}

		// Admin routes (JWT protected)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Admin quota info
			admin.GET("/quota", adminHandler.GetAdminQuota)

			// Room management
			rooms := admin.Group("/rooms")
			{
				rooms.POST("", roomHandler.CreateRoom)
				rooms.GET("", roomHandler.ListRooms)
				rooms.GET("/:id", roomHandler.GetRoom)
				rooms.PUT("/:id", roomHandler.UpdateRoom)
				rooms.DELETE("/:id", roomHandler.DeleteRoom)

				// Realtime monitoring (must use same parameter name)
				rooms.GET("/:id/realtime", votingHandler.GetRealtimeVoteData)
			}

			// Candidate management
			candidates := admin.Group("/candidates")
			{
				candidates.POST("", candidateHandler.CreateCandidate)
				candidates.GET("/:id", candidateHandler.GetCandidate)
				candidates.PUT("/:id", candidateHandler.UpdateCandidate)
				candidates.DELETE("/:id", candidateHandler.DeleteCandidate)
				candidates.GET("/room/:roomId", candidateHandler.ListCandidatesByRoom)
			}

			// Ticket management
			tickets := admin.Group("/tickets")
			{
				tickets.POST("", ticketHandler.CreateTicket)
				tickets.POST("/bulk", ticketHandler.CreateTicketsBulk)
				tickets.GET("/room/:roomId", ticketHandler.ListTicketsByRoom)
				tickets.DELETE("/:id", ticketHandler.DeleteTicket)
			}
		}
	}

	return router
}
