package heartbeat_mock

// Function to create heartbeat (fake implementation)
func CreateHeartbeat(authToken string, jsonData string) (string, error) {
	// Return a fake URL
	fakeURL := "https://fakeurl.com/heartbeat"
	return fakeURL, nil
}
