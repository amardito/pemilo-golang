package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/amard/pemilo-golang/internal/config"
	"github.com/amard/pemilo-golang/internal/handler"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/repository"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/amard/pemilo-golang/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env (ignore error if not found — production uses real env vars)
	godotenv.Load()

	cfg := config.Load()

	// Database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	// Run migrations automatically
	if err := util.RunMigrations(db); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	eventRepo := repository.NewEventRepo(db)
	slateRepo := repository.NewSlateRepo(db)
	voterRepo := repository.NewVoterRepo(db)
	voterTokenRepo := repository.NewVoterTokenRepo(db)
	ballotRepo := repository.NewBallotRepo(db)
	auditLogRepo := repository.NewAuditLogRepo(db)
	orderRepo := repository.NewOrderRepo(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg)
	eventService := service.NewEventService(eventRepo, auditLogRepo)
	slateService := service.NewSlateService(slateRepo, eventRepo)
	voterService := service.NewVoterService(voterRepo, voterTokenRepo, eventRepo, auditLogRepo)
	voteService := service.NewVoteService(db, eventRepo, slateRepo, voterRepo, voterTokenRepo, ballotRepo)
	statsService := service.NewStatsService(ballotRepo, eventRepo)
	auditService := service.NewAuditService(auditLogRepo, eventRepo)
	paymentService := service.NewPaymentService(orderRepo, eventRepo, cfg)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	eventHandler := handler.NewEventHandler(eventService)
	slateHandler := handler.NewSlateHandler(slateService)
	voterHandler := handler.NewVoterHandler(voterService)
	votePublicHandler := handler.NewVotePublicHandler(voteService)
	statsHandler := handler.NewStatsHandler(statsService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	auditLogHandler := handler.NewAuditLogHandler(auditService)

	// Rate limiters
	votePrepareRL := middleware.NewRateLimiter(20, 5) // 20 req/min, burst 5
	voteSubmitRL := middleware.NewRateLimiter(20, 1)  // 20 req/min, burst 1

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware("*"))
	r.Use(middleware.ErrorLogger())

	api := r.Group("/api")
	{
		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.Me)
		}

		// Protected admin routes
		admin := api.Group("")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Events
			admin.POST("/events", eventHandler.Create)
			admin.GET("/events", eventHandler.List)
			admin.GET("/events/:eventId", eventHandler.Get)
			admin.PATCH("/events/:eventId", eventHandler.Update)
			admin.POST("/events/:eventId/open", eventHandler.Open)
			admin.POST("/events/:eventId/close", eventHandler.Close)
			admin.POST("/events/:eventId/lock", eventHandler.Lock)

			// Slates
			admin.POST("/events/:eventId/slates", slateHandler.Create)
			admin.GET("/events/:eventId/slates", slateHandler.List)
			admin.PATCH("/slates/:slateId", slateHandler.Update)
			admin.DELETE("/slates/:slateId", slateHandler.Delete)

			// Slate Members
			admin.POST("/slates/:slateId/members", slateHandler.CreateMember)
			admin.PATCH("/slate-members/:memberId", slateHandler.UpdateMember)
			admin.DELETE("/slate-members/:memberId", slateHandler.DeleteMember)

			// Voters
			admin.POST("/events/:eventId/voters/import", voterHandler.Import)
			admin.GET("/events/:eventId/voters", voterHandler.List)
			admin.POST("/events/:eventId/voters/tokens/generate", voterHandler.GenerateTokens)
			admin.GET("/events/:eventId/voters/tokens/export", voterHandler.ExportTokens)
			admin.GET("/events/:eventId/voters/turnout/export", voterHandler.ExportTurnout)
			admin.GET("/voters/template", voterHandler.DownloadTemplate)

			// Stats
			admin.GET("/events/:eventId/stats", statsHandler.GetStats)

			// Audit Logs
			admin.GET("/events/:eventId/audit-logs", auditLogHandler.List)

			// Payment
			admin.POST("/events/:eventId/upgrade", paymentHandler.Upgrade)
			admin.GET("/orders/:orderId", paymentHandler.GetOrder)
		}

		// Public voting routes (no auth, rate-limited)
		public := api.Group("/public")
		{
			public.POST("/events/:eventId/vote/prepare", votePrepareRL.Middleware(), votePublicHandler.Prepare)
			public.POST("/events/:eventId/vote/submit", voteSubmitRL.Middleware(), votePublicHandler.Submit)
		}

		// Payment webhook (no auth, verified by signature)
		api.POST("/payments/ipaymu/webhook", paymentHandler.Webhook)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
