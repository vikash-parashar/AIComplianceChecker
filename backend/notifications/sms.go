package notifications

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SendSMS hooks into Twilio logic using pure HTTP
func SendSMS(to string, message string) error {
	accountSid := os.Getenv("TWILIO_SID")
	authToken := os.Getenv("TWILIO_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM")

	if accountSid == "" || authToken == "" {
		log.Printf("[MOCK SMS] To: %s | Message: %s", to, message)
		return nil
	}

	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", fromNumber)
	data.Set("Body", message)

	req, err := http.NewRequest("POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(accountSid+":"+authToken))
	req.Header.Add("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send SMS API req: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Twilio Error %d: %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("Twilio error")
	}

	return nil
}

func SendPaymentFailedSMS(to string) {
	if to == "" {
		return
	}
	SendSMS(to, "AI Compliance: Your requested payment failed to process. Please check your dashboard for details.")
}
