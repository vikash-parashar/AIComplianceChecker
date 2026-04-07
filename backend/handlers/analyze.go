package handlers

import (
	"context"
	"net/http"

	"github.com/AIComplianceChecker/backend/database"
	"github.com/AIComplianceChecker/backend/engine"
	"github.com/AIComplianceChecker/backend/models"
	"github.com/AIComplianceChecker/backend/notifications"
	"github.com/gin-gonic/gin"
)

func AnalyzeSMS(c *gin.Context) {
	analyzeCommon(c, "sms")
}

func AnalyzePolicy(c *gin.Context) {
	analyzeCommon(c, "policy")
}

func AnalyzeConfig(c *gin.Context) {
	// For config, we extract string but treat as config logic
	analyzeCommon(c, "config")
}

func analyzeCommon(c *gin.Context, checkType string) {
	userIDStr := c.GetString("userID")

	var req models.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// try config struct
		var cfreq models.ConfigAnalyzeRequest
		if err2 := c.ShouldBindJSON(&cfreq); err2 == nil {
			req.Content = cfreq.ConfigJSON
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content provided"})
			return
		}
	}

	// 0. Verify usage credits
	var credits int
	err := database.Pool.QueryRow(context.Background(), "SELECT credits FROM users WHERE id = $1", userIDStr).Scan(&credits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify credits"})
		return
	}
	if credits <= 0 {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient credits to perform analysis."})
		return
	}

	// 1. Create check log entry
	var checkID string
	err = database.Pool.QueryRow(context.Background(),
		"INSERT INTO checks (user_id, check_type, status) VALUES ($1, $2, $3) RETURNING id",
		userIDStr, checkType, "pending",
	).Scan(&checkID)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logic check in DB"})
		return
	}

	var allViolations []models.Violation

	// 2. Static Rules (Only applied loosely based on type, e.g. SMS)
	if checkType == "sms" {
		allViolations = engine.CheckA2P10DLC(req.Content)
	}

	// 3. AI Rules
	aiResult, err := engine.CheckWithAI(req.Content, checkType)
	if err == nil && aiResult != nil {
		for _, v := range aiResult.Violations {
			allViolations = append(allViolations, models.Violation{
				Issue:    v.Issue,
				Severity: v.Severity,
				Fix:      v.Fix,
				RuleType: v.RuleType,
			})
		}
	}

	// 4. Save Violations
	for _, v := range allViolations {
		_, _ = database.Pool.Exec(context.Background(),
			"INSERT INTO violations (check_id, issue, severity, fix, rule_type) VALUES ($1, $2, $3, $4, $5)",
			checkID, v.Issue, v.Severity, v.Fix, v.RuleType,
		)
	}

	// Deduct a credit securely
	_, _ = database.Pool.Exec(context.Background(), "UPDATE users SET credits = credits - 1 WHERE id = $1", userIDStr)

	// Low credit notification threshold (send async tracking email if exactly 2 left!)
	if credits-1 == 2 {
		var email string
		_ = database.Pool.QueryRow(context.Background(), "SELECT email FROM users WHERE id = $1", userIDStr).Scan(&email)
		if email != "" {
			go notifications.SendLowCreditWarning(email, 2)
		}
	}

	// Update status
	_, _ = database.Pool.Exec(context.Background(), "UPDATE checks SET status = 'completed' WHERE id = $1", checkID)

	c.JSON(http.StatusOK, gin.H{
		"check_id":   checkID,
		"violations": allViolations,
	})
}

func GetChecks(c *gin.Context) {
	userIDStr := c.GetString("userID")

	rows, err := database.Pool.Query(context.Background(),
		"SELECT id, check_type, status, created_at FROM checks WHERE user_id = $1 ORDER BY created_at DESC",
		userIDStr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve checks"})
		return
	}
	defer rows.Close()

	var checks []models.Check
	for rows.Next() {
		var check models.Check
		if err := rows.Scan(&check.ID, &check.CheckType, &check.Status, &check.CreatedAt); err != nil {
			continue
		}
		checks = append(checks, check)
	}

	c.JSON(http.StatusOK, gin.H{"checks": checks})
}
