package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetGitHubToken retrieves GitHub token from multiple sources in order of priority:
// 1. Viper configuration (passed as parameter)
// 2. GITHUB_TOKEN environment variable
// 3. gh CLI token if authenticated
func GetGitHubToken(configToken string) (string, error) {
	// Check viper config first
	if configToken != "" {
		return configToken, nil
	}

	// Check environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Check gh CLI authentication
	token, err := getGHCLIToken()
	if err == nil && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("GitHub token not found. Please use one of the following methods:\n" +
		"1. Set GITHUB_TOKEN environment variable\n" +
		"2. Configure token in .ghsecrets.yaml\n" +
		"3. Login with 'gh auth login'")
}

// getGHCLIToken retrieves the token from gh CLI if authenticated
func getGHCLIToken() (string, error) {
	// First check if gh is authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh CLI not authenticated")
	}

	// Get the token
	cmd = exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get gh token: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("gh token is empty")
	}

	return token, nil
}