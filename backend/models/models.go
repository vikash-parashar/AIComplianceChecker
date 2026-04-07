package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	Credits          int       `json:"credits"`
	StripeCustomerID string    `json:"stripe_customer_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Check represents a compliance check log
type Check struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CheckType string    `json:"check_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Violation represents an issue found during a compliance check
type Violation struct {
	ID        uuid.UUID `json:"id"`
	CheckID   uuid.UUID `json:"check_id"`
	Issue     string    `json:"issue"`
	Severity  string    `json:"severity"`
	Fix       string    `json:"fix"`
	RuleType  string    `json:"rule_type"`
}

// API Requests

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AnalyzeRequest struct {
	Content string `json:"content" binding:"required"`
}

type ConfigAnalyzeRequest struct {
	ConfigJSON string `json:"config_json" binding:"required"`
}

// ComplianceResult is the structured JSON we want from OpenAI
type ComplianceResult struct {
	Violations []struct {
		Issue    string `json:"issue"`
		Severity string `json:"severity"` // low, medium, high, critical
		Fix      string `json:"fix"`
		RuleType string `json:"rule_type"` // hipaa, gdpr, a2p
	} `json:"violations"`
}
