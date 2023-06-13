package heartbeat_mock

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

// Function to create heartbeat (fake implementation)
func CreateHeartbeat(authToken string, jsonData string) (string, error) {
	// Generate a random byte slice
	bytes := make([]byte, 10) // length of the random string
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	// Convert the byte slice to a string
	randomString := hex.EncodeToString(bytes)

	// Return a fake URL with the random string appended
	fakeURL := "https://uptime.betterstack.fake.com/heartbeat/" + randomString
	return fakeURL, nil
}

func CreateHeartbeatGroup(authToken string, heartbeatGroupName string) (string, error) {
	rand.Seed(time.Now().UnixNano())

	// Generate a random 3-digit integer
	heartbeatGroupID := fmt.Sprintf("%03d", rand.Intn(1000))

	return heartbeatGroupID, nil
}

func GetHeartbeatGroupID(authToken string, heartbeatGroupName string) string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random 3-digit integer
	heartbeatGroupID := fmt.Sprintf("%03d", rand.Intn(1000))

	return heartbeatGroupID
}
