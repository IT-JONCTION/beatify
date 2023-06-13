package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/robfig/cron"
	"golang.org/x/time/rate"
)

type HeartbeatResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			URL              string `json:"url"`
			Name             string `json:"name"`
			Period           int    `json:"period"`
			Grace            int    `json:"grace"`
			Call             bool   `json:"call"`
			SMS              bool   `json:"sms"`
			Email            bool   `json:"email"`
			Push             bool   `json:"push"`
			TeamWait         int    `json:"team_wait"`
			HeartbeatGroupID int    `json:"heartbeat_group_id"`
			SortIndex        int    `json:"sort_index"`
			PausedAt         string `json:"paused_at"`
			CreatedAt        string `json:"created_at"`
			UpdatedAt        string `json:"updated_at"`
		} `json:"attributes"`
	} `json:"data"`
}

type HeartbeatGroupResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name      string `json:"name"`
			SortIndex int    `json:"sort_index"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Paused    bool   `json:"paused"`
		} `json:"attributes"`
	} `json:"data"`
}

type Pagination struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

type HeartbeatGroupsResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name      string `json:"name"`
			SortIndex int    `json:"sort_index"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Paused    bool   `json:"paused"`
		} `json:"attributes"`
	} `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func GetHeartbeatGroupID(authToken string, heartbeatGroupName string) (string, error) {
	limiter := rate.NewLimiter(1, 5) // Limit to 1 request per second, with bursts up to 5 requests.
	url := "https://uptime.betterstack.com/api/v2/heartbeat-groups"

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("Error creating HTTP request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{}
		err = limiter.Wait(context.Background()) // Wait for the limiter to allow us to make the request.
		if err != nil {
			return "", fmt.Errorf("Error waiting for rate limiter: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("Error sending HTTP request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return "", fmt.Errorf("Error response status code: %d", resp.StatusCode)
		}

		// If the status is StatusNotFound, return an empty string without an error
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Error reading response body: %w", err)
		}

		var response HeartbeatGroupsResponse
		err = json.Unmarshal(responseBody, &response)
		if err != nil {
			return "", fmt.Errorf("Error unmarshalling response body: %w", err)
		}

		for _, data := range response.Data {
			if data.Attributes.Name == heartbeatGroupName {
				return data.ID, nil
			}
		}

		if response.Pagination.Next == "" {
			break
		}
		url = response.Pagination.Next
		time.Sleep(1 * time.Second) // Sleep to respect the rate limit.
	}

	return "", nil
}

func CreateHeartbeatGroup(authToken string, heartbeatGroupName string) (string, error) {
	url := "https://uptime.betterstack.com/api/v2/heartbeat-groups"

	// Define the data to send in the request body
	data := map[string]string{
		"name": heartbeatGroupName,
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Error creating JSON request body: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
		return "", fmt.Errorf("Error response status code: %d", resp.StatusCode)
	}

	// Read the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %w", err)
	}

	// Extract the GroupId from the response body
	heartbeatGroupID, err := extractHeartbeatGroupIDFromResponse(responseBody, heartbeatGroupName)
	if err != nil {
		return "", fmt.Errorf("Error extracting GroupId from response: %w", err)
	}

	return heartbeatGroupID, nil
}

func extractHeartbeatGroupIDFromResponse(responseBody []byte, heartbeatGroupName string) (string, error) {
	var heartbeatGroupResp HeartbeatGroupResponse
	err := json.Unmarshal(responseBody, &heartbeatGroupResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal Heartbeat Groups response body: %w", err)
	}
	if heartbeatGroupResp.Data.Attributes.Name == heartbeatGroupName {
		return heartbeatGroupResp.Data.ID, nil
	}
	return "", fmt.Errorf("heartbeat group not found")
}

// Utility function to extract URL from response body
func extractURLFromResponse(responseBody []byte) (string, error) {
	var heartbeatResp HeartbeatResponse
	err := json.Unmarshal(responseBody, &heartbeatResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal Heartbeat response body: %w", err)
	}

	return heartbeatResp.Data.Attributes.URL, nil
}

// Function to prepare config JSON
func PrepareConfigJson(crontab, heartbeatName string, heartbeatGroupID string) (string, error) {
	// Create a new cron parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// Parse the crontab schedule
	schedule, err := parser.Parse(crontab)
	if err != nil {
		return "", fmt.Errorf("Error parsing crontab schedule: %w", err)
	}

	// Get the next scheduled time from current time
	nextRun := schedule.Next(time.Now())

	// Get the next scheduled time from the next run
	nextNextRun := schedule.Next(nextRun)

	// Calculate the period in seconds between two runs
	period := int(nextNextRun.Sub(nextRun).Seconds())

	// Calculate the grace period as approximately 20% of the period
	grace := int(float64(period) * 0.2)

	// Create the JSON representation
	jsonData := fmt.Sprintf(`{
		"name": "%s",
		"period": %d,
		"grace": %d`,
		heartbeatName, period, grace)

	// Include heartbeat_group_id if provided
	if heartbeatGroupID != "" {
		jsonData += fmt.Sprintf(`,
		"heartbeat_group_id": "%s"`, heartbeatGroupID)
	}

	jsonData += "\n}"

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
