package heartbeat_mock

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
