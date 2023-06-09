package heartbeat

import (
	"fmt"
	"time"
	"github.com/robfig/cron"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
)

type HeartbeatResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			URL                string `json:"url"`
			Name               string `json:"name"`
			Period             int    `json:"period"`
			Grace              int    `json:"grace"`
			Call               bool   `json:"call"`
			SMS                bool   `json:"sms"`
			Email              bool   `json:"email"`
			Push               bool   `json:"push"`
			TeamWait           int    `json:"team_wait"`
			HeartbeatGroupID   string `json:"heartbeat_group_id"`
			SortIndex          int    `json:"sort_index"`
			PausedAt           string `json:"paused_at"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
		} `json:"attributes"`
	} `json:"data"`
}


// Utility function to extract URL from response body
func extractURLFromResponse(responseBody []byte) (string, error) {
	var heartbeatResp HeartbeatResponse
	err := json.Unmarshal(responseBody, &heartbeatResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return heartbeatResp.Data.Attributes.URL, nil
}

// Function to curl API providing Auth Token and Heartbeat Name and config json
func CurlAPI() error {
	fmt.Println("curling API")
	return nil
}

// Function to prepare config JSON
func PrepareConfigJson(crontab, heartbeatName string) (string, error) {
	// Create a new cron parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// Parse the crontab schedule
	schedule, err := parser.Parse(crontab)
	if err != nil {
		return "", fmt.Errorf("Error parsing crontab schedule: %w", err)
	}

	// Calculate the period in seconds
	period := int(schedule.Next(time.Now()).Sub(time.Now()).Seconds())

	// Calculate the grace period as approximately 20% of the period
	grace := int(float64(period) * 0.2)

	// Create the JSON representation
	jsonData := fmt.Sprintf(`{
		"name": "%s",
		"period": %d,
		"grace": %d
	}`, heartbeatName, period, grace)

	return jsonData, nil
}

// Function to create heartbeat
func CreateHeartbeat(authToken string, jsonData string) (string, error) {
	url := "https://uptime.betterstack.com/api/v2/heartbeats"

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(jsonData))
	if err != nil {
		return "", fmt.Errorf("Error creating HTTP request: %w", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP client and send the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Unexpected response status: %s", resp.Status)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %w", err)
	}

	// Extract the URL from the response using the utility function
	resultURL, err := extractURLFromResponse(responseBody)
	if err != nil {
		return "", err
	}

	return resultURL, nil
}
