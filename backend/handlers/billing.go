package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/AIComplianceChecker/backend/database"
	"github.com/AIComplianceChecker/backend/notifications"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
)

func GetUserCredits(c *gin.Context) {
	userIDStr := c.GetString("userID")

	var credits int
	err := database.Pool.QueryRow(context.Background(), "SELECT credits FROM users WHERE id = $1", userIDStr).Scan(&credits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failure to load credit balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credits": credits})
}

func Checkout(c *gin.Context) {
	userIDStr := c.GetString("userID")
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// If stripe isn't configured, fallback to mock adding credits instantly for Local Dev Testing
	if stripe.Key == "" {
		_, err := database.Pool.Exec(context.Background(), "UPDATE users SET credits = credits + 10 WHERE id = $1", userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"checkout_url": "mock_success", 
			"message":      "[MOCK MODE] Added 10 credits successfully since keys are missing!",
		})
		return
	}

	domain := "http://localhost:3000" // Should map to env var in prod

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("100 Compliance AI Credits"),
					},
					UnitAmount: stripe.Int64(2000), // $20.00
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "?status=success"),
		CancelURL:  stripe.String(domain + "?status=canceled"),
		ClientReferenceID: stripe.String(userIDStr),
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": s.URL, 
		"message":      "Redirecting to Stripe...",
	})
}

// StripeWebhook is an unauthenticated endpoint that Stripe talks to securely
func StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Body read error"})
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Verify the signature securely
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature mismatch"})
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format"})
			return
		}

		userIDStr := session.ClientReferenceID
		
		// 1. Credit the user with 100 Credits securely
		_, err = database.Pool.Exec(context.Background(), "UPDATE users SET credits = credits + 100 WHERE id = $1", userIDStr)
		if err != nil {
			log.Printf("Failed to increment user DB %s: %v", userIDStr, err)
			return
		}

		// 2. Dispatch the Email Receipt
		var email string
		_ = database.Pool.QueryRow(context.Background(), "SELECT email FROM users WHERE id = $1", userIDStr).Scan(&email)
		if email != "" {
			go notifications.SendPaymentSuccess(email, 100)
		}
	} else if event.Type == "payment_intent.payment_failed" {
		// Mock logic: send an SMS if phone exists, fallback to email
		// Actually, let's just log it since we don't know the user cleanly from payment_intent yet, 
		// but typically you'd link it from the stripe customer.
		log.Printf("Payment failed intent encountered")
	}

	c.JSON(http.StatusOK, gin.H{"status": "handled"})
}
