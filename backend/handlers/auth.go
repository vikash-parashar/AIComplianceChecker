package handlers

import (
	"context"
	"net/http"

	"github.com/AIComplianceChecker/backend/auth"
	"github.com/AIComplianceChecker/backend/database"
	"github.com/AIComplianceChecker/backend/models"
	"github.com/AIComplianceChecker/backend/notifications"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
		return
	}

	var newUserID string
	err = database.Pool.QueryRow(context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		req.Email, hash,
	).Scan(&newUserID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists or DB error"})
		return
	}

	token, err := auth.GenerateToken(newUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Dispatch Asynchronous Lifecycle Welcome Email
	go notifications.SendWelcomeEmail(req.Email)

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":    newUserID,
			"email": req.Email,
		},
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.User
	err := database.Pool.QueryRow(context.Background(),
		"SELECT id, email, password_hash FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}
