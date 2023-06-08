package heartbeat

import (
	"fmt"
	"time"
	"github.com/robfig/cron"
)

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
