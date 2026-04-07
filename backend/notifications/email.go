package notifications

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// SendEmail is a generic wrapper to dispatch transactional emails
func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST") // e.g. smtp.sendgrid.net
	port := os.Getenv("SMTP_PORT") // e.g. 587

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	// If SMTP is not configured, we gracefully degrade to stdout logging for dev testability
	if host == "" {
		log.Printf("[MOCK EMAIL] To: %s | Subject: %s | Body: %s", to, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	return nil
}

func SendWelcomeEmail(to string) {
	subject := "Welcome to AI Compliance Checker!"
	body := "Thanks for joining. You have been credited with 10 free scans. If you need more usage or have any issues, reply to this email!"
	SendEmail(to, subject, body)
}

func SendLowCreditWarning(to string, credits int) {
	subject := "Low Credits Warning!"
	body := fmt.Sprintf("You currently have only %d credits remaining for compliance checking. Please login and purchase more credits to ensure uninterrupted service.", credits)
	SendEmail(to, subject, body)
}

func SendPaymentSuccess(to string, creditsAdded int) {
	subject := "Payment Successful!"
	body := fmt.Sprintf("Thank you for your top-up. We have successfully added %d credits to your account. They are ready to use immediately.", creditsAdded)
	SendEmail(to, subject, body)
}
