package main

import (
	"log"

	"github.com/AIComplianceChecker/backend/database"
	"github.com/AIComplianceChecker/backend/handlers"
	"github.com/AIComplianceChecker/backend/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Ignore error, as env vars might be set in system

	err := database.ConnectDB()
	if err != nil {
		log.Printf("Warning: Failed to connect to DB on startup: %v. Database functionality may be limited unless started.", err)
	} else {
		err = database.Migrate()
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		defer database.CloseDB()
	}

	router := gin.Default()

	// Setup CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Adjust for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Protected analyze routes
	analyze := router.Group("/api/analyze")
	analyze.Use(middleware.AuthMiddleware())
	{
		analyze.POST("/sms", handlers.AnalyzeSMS)
		analyze.POST("/policy", handlers.AnalyzePolicy)
		analyze.POST("/config", handlers.AnalyzeConfig)
	}

	router.GET("/api/checks", middleware.AuthMiddleware(), handlers.GetChecks)
	
	// Open public endpoint for Stripe to securely ping when checkout completes!
	router.POST("/api/webhook/stripe", handlers.StripeWebhook)

	billing := router.Group("/api/billing")
	billing.Use(middleware.AuthMiddleware())
	{
		billing.GET("/credits", handlers.GetUserCredits)
		billing.POST("/checkout", handlers.Checkout)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("Starting AI Compliance Checker API on port 8085...")
	if err := router.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
