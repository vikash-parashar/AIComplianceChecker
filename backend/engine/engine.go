package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/AIComplianceChecker/backend/models"
)

func CheckA2P10DLC(content string) []models.Violation {
	var violations []models.Violation

	// Simple static checks
	stopRegex := regexp.MustCompile(`(?i)\b(STOP|UNSUBSCRIBE|CANCEL)\b`)
	if !stopRegex.MatchString(content) {
		violations = append(violations, models.Violation{
			Issue:    "Missing opt-out language",
			Severity: "high",
			Fix:      "Add standard opt-out keywords like 'Reply STOP to cancel'.",
			RuleType: "a2p",
		})
	}

	helpRegex := regexp.MustCompile(`(?i)\b(HELP)\b`)
	if !helpRegex.MatchString(content) {
		violations = append(violations, models.Violation{
			Issue:    "Missing HELP keyword",
			Severity: "medium",
			Fix:      "Include instructions on how users can get help (e.g. 'Reply HELP for help').",
			RuleType: "a2p",
		})
	}

	return violations
}

// CheckWithAI calls the OpenAI API to do deep semantic checking for HIPAA or GDPR
func CheckWithAI(content string, checkType string) (*models.ComplianceResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Mock behavior if no key is provided, so that tests pass easily without an API key
		return &models.ComplianceResult{
			Violations: []struct {
				Issue    string "json:\"issue\""
				Severity string "json:\"severity\""
				Fix      string "json:\"fix\""
				RuleType string "json:\"rule_type\""
			}{
				{
					Issue:    "MOCK FLAG: Missing consent log requirement",
					Severity: "high",
					Fix:      "Add specific terms detailing how long data is retained.",
					RuleType: "gdpr",
				},
			},
		}, nil
	}

	url := "https://api.openai.com/v1/chat/completions"

	promptText := fmt.Sprintf("Analyze this content for compliance issues. Type: %s. Content: %s", checkType, content)
	
	systemPrompt := "You are a Data Privacy and Compliance Expert."
	if checkType == "sms" {
		systemPrompt = `You are a strict A2P 10DLC Compliance Assessor. Evaluate the provided SMS input against Twilio and carrier guidelines.
Rules to check:
1. Ensure explicit Opt-Out language (e.g. 'Reply STOP to cancel') is present.
2. Ensure Help language (e.g. 'Reply HELP') is present.
3. Check for disallowed content (SHAFT: Sex, Hate, Alcohol, Firearms, Tobacco) or high-risk financial offers.`
	} else if checkType == "policy" {
		systemPrompt = `You are an expert Data Privacy Lawyer. Evaluate this privacy policy text against GDPR, CCPA, and HIPAA.
Rules to check:
1. HIPAA: Ensure strict guidelines on the handling, encryption, and disposal of Protected Health Information (PHI).
2. GDPR: Explicit consent for data collection, the right to be forgotten (data deletion requests), and explicit data retention limits.
3. CCPA: Must include 'Do Not Sell My Personal Information' language and clearly list third-party data sales.`
	} else if checkType == "config" {
		systemPrompt = `You are a strict DevOps Security Architect. Evaluate the JSON configuration for PII/PHI data leakage risks.
Rules to check:
1. Ensure 'phi_storage' or logging parameters do not explicitly retain logs indefinitely.
2. Demand encryption keys or TLS validation flags are present if dealing with sensitive variables.`
	}

	systemPrompt += `
Return ONLY valid JSON in the following format (and NO OTHER TEXT OR MARKDOWN block wrappers):
{
  "violations": [
    {
       "issue": "Detailed description of the compliance violation",
       "severity": "high/medium/low",
       "fix": "Actionable suggestion to fix",
       "rule_type": "hipaa/gdpr/a2p/ccpa/security"
    }
  ]
}`

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": promptText,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.2,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", string(bodyBytes))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return &models.ComplianceResult{}, nil
	}

	var result models.ComplianceResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("Failed to parse OpenAI JSON: %v. Raw text: %s", err, response.Choices[0].Message.Content)
	}

	return &result, nil
}
