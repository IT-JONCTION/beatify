package config

import (
	"fmt"
	"strings"
)

func PromptAuthToken() string {
	var authToken string

	fmt.Print("Enter the authentication token for the BetterUptime API: ")
	fmt.Scanln(&authToken)
	fmt.Println()

	// Trim any leading or trailing whitespace from the input
	authToken = strings.TrimSpace(authToken)

	return authToken
}